//! Contains the parsed request in `context` as well as various
//! helper methods to ensure the request is handled correctly,
//! such as reading the body only once.
const Request = @This();

const root = @import("root");
const std = @import("std");
const Uri = @import("Uri.zig");
const Allocator = std.mem.Allocator;
const mem = std.mem;
const Stream = std.net.Stream;

const max_buffer_size = blk: {
    const given = if (@hasDecl(root, "buffer_size")) root.buffer_size else 1024 * 64; // 64kB
    break :blk std.math.min(given, 1024 * 1024 * 16); // max stack size
};

/// Internal allocator, fed by an arena allocator. Any memory allocated using this
/// allocator will be freed upon the end of a request. It's therefore illegal behaviour
/// to read from/write to anything allocated with this after a request and must be duplicated first,
/// or allocated using a different strategy.
arena: Allocator,
/// Context provides all information from the actual request that was made by the client.
context: Context,

/// Alias to Stream.ReadError
/// Possible errors that can occur when reading from the connection stream.
pub const ReadError = Stream.ReadError;

/// HTTP methods as specified in RFC 7231
pub const Method = enum {
    get,
    head,
    post,
    put,
    delete,
    any,

    fn fromString(string: []const u8) Method {
        return if (std.mem.eql(u8, string, "GET"))
            .get
        else if (std.mem.eql(u8, string, "POST"))
            .post
        else if (std.mem.eql(u8, string, "PUT"))
            .put
        else if (std.mem.eql(u8, string, "DELETE"))
            .delete
        else if (std.mem.eql(u8, string, "HEAD"))
            @as(Method, .head)
        else
            .any;
    }
};

/// Represents an HTTP Header
pub const Header = struct {
    key: []const u8,
    value: []const u8,
};

/// Alias to StringHashMapUnmanaged([]const u8)
pub const Headers = std.StringHashMapUnmanaged([]const u8);
/// Buffered reader for reading the connection stream
pub const Reader = std.io.BufferedReader(4096, Stream.Reader).Reader;

/// `Context` contains the result from the parser.
/// `Request` uses this information to handle correctness when parsing
/// a body or the headers.
pub const Context = struct {
    method: Method,
    /// URI object, contains the parsed and validated path and optional
    /// query and fragment.
    /// Note: For http 1.1 this may also contain the authority component.
    uri: Uri,
    /// HTTP Request headers data.
    raw_header_data: []const u8,
    /// Hostname the request was sent to. Includes its port. Required for HTTP/1.1
    host: ?[]const u8,
    /// Body of the request. Its livetime equals that of the request itself,
    /// meaning that any access to its data beyond that is illegal and must be duplicated
    /// to extend its lifetime.
    raw_body: []const u8,
    /// State of the connection. `keep_alive` is the default for HTTP/1.1 and `close` for earlier versions.
    /// For HTTP/2.2 browsers such as Chrome and Firefox ignore this.
    connection_type: enum {
        keep_alive,
        close,
    },
};

/// Iterator to iterate through headers
const Iterator = struct {
    slice: []const u8,
    index: usize,

    /// Searches for the next header.
    /// Parsing cannot be failed as that would have been caught by `parse()`
    pub fn next(self: *Iterator) ?Header {
        if (self.index >= self.slice.len) return null;

        var state: enum { key, value } = .key;

        var header: Header = undefined;
        var start = self.index;
        while (self.index < self.slice.len) : (self.index += 1) {
            const c = self.slice[self.index];
            if (state == .key and c == ':') {
                header.key = self.slice[start..self.index];
                start = self.index + 2;
                state = .value;
            }
            if (state == .value and c == '\r') {
                header.value = self.slice[start..self.index];
                self.index += 2;
                return header;
            }
        }

        return null;
    }
};

/// Creates an iterator to retrieve all headers
/// As the data is known, this does not require any allocations
/// If all headers needs to be known at once, use `headers()`.
pub fn iterator(self: Request) Iterator {
    return Iterator{
        .slice = self.context.raw_header_data[0..],
        .index = 0,
    };
}

/// Creates an unmanaged Hashmap from the request headers, memory is owned by caller
/// Every header key and value will be allocated for the map and must therefore be freed
/// manually as well.
pub fn headers(self: Request, gpa: Allocator) !Headers {
    var map = Headers{};

    var it = self.iterator();
    while (it.next()) |header| {
        try map.put(gpa, try gpa.dupe(u8, header.key), try gpa.dupe(u8, header.value));
    }

    return map;
}

/// Returns the content of the body
/// Its livetime equals that of the request itself,
/// meaning that any access to its data beyond that is illegal and must be duplicated
/// to extend its lifetime.
pub fn body(self: Request) []const u8 {
    return self.context.raw_body;
}

/// Returns the path of the request
/// To retrieve the raw path, access `context.uri.raw_path`
pub fn path(self: Request) []const u8 {
    return self.context.uri.path;
}

/// Returns the method of the request as `Method`
pub fn method(self: Request) Method {
    return self.context.method;
}

/// Returns the host. This cannot be null when the request
/// is HTTP 1.1 or higher.
pub fn host(self: Request) ?[]const u8 {
    return self.context.host;
}

/// Errors which can occur during the parsing of
/// a HTTP request.
pub const ParseError = error{
    OutOfMemory,
    /// Method is missing or invalid
    InvalidMethod,
    /// URL is missing in status line or invalid
    InvalidUrl,
    /// Protocol in status line is missing or invalid
    InvalidProtocol,
    /// Headers are missing
    MissingHeaders,
    /// Invalid header was found
    IncorrectHeader,
    /// Buffer overflow when parsing an integer
    Overflow,
    /// Invalid character when parsing an integer
    InvalidCharacter,
    /// When the connection has been closed or no more data is available
    EndOfStream,
    /// Provided request's size is bigger than max size (2^32).
    StreamTooLong,
    /// Request headers are too large and do not find in `buffer_size`
    HeadersTooLarge,
    /// Line ending of the requests are corrupted/invalid. According to the http
    /// spec, each line must end with \r\n
    InvalidLineEnding,
    /// When body is incomplete
    InvalidBody,
    /// When the client uses HTTP version 1.1 and misses the 'host' header, this error will
    /// be returned.
    MissingHost,
};

/// Parse accepts a `Reader`. It will read all data it contains
/// and tries to parse it into a `Request`. Can return `ParseError` if data is corrupt.
/// The Allocator is made available to users that require an allocation for during a single request,
/// as an arena is passed in by the `Server`. The provided buffer is used to parse the actual content,
/// meaning the entire request -> response can be done with no allocations.
pub fn parse(gpa: Allocator, reader: anytype, buffer: []u8) (ParseError || Stream.ReadError)!Request {
    return Request{
        .arena = gpa,
        .context = try parseContext(
            gpa,
            reader,
            buffer,
        ),
    };
}

fn parseContext(gpa: Allocator, reader: anytype, buffer: []u8) (ParseError || @TypeOf(reader).Error)!Context {
    var ctx: Context = .{
        .method = .get,
        .uri = Uri.empty,
        .raw_header_data = undefined,
        .host = null,
        .raw_body = "",
        .connection_type = .keep_alive,
    };

    var parser = Parser(@TypeOf(reader)).init(gpa, buffer, reader);
    while (try parser.nextEvent()) |event| {
        switch (event) {
            .status => |status| {
                ctx.connection_type = .keep_alive;
                ctx.uri = Uri.parse(status.path) catch return error.InvalidUrl;
                ctx.method = Request.Method.fromString(status.method);
            },
            .header => |header| {
                if (ctx.connection_type == .keep_alive and
                    std.ascii.eqlIgnoreCase(header.key, "connection"))
                {
                    if (std.ascii.eqlIgnoreCase(header.value, "close")) ctx.connection_type = .close;
                }

                if (ctx.host == null and std.ascii.eqlIgnoreCase(header.key, "host")) {
                    ctx.host = header.value;
                    _ = Uri.parseAuthority(&ctx.uri, header.value) catch return error.IncorrectHeader;
                }
            },
            .end_of_header => ctx.raw_header_data = buffer[parser.header_start..parser.header_end],
            .body => |content| ctx.raw_body = content,
        }
    }

    if (ctx.host == null) return error.MissingHost;

    return ctx;
}

fn Parser(ReaderType: anytype) type {
    return struct {
        const Self = @This();

        gpa: Allocator,
        buffer: []u8,
        index: usize,
        state: std.meta.Tag(Event),
        reader: ReaderType,
        done: bool,
        content_length: usize,
        header_start: usize,
        header_end: usize,
        chunked: bool,

        const Event = union(enum) {
            status: struct {
                method: []const u8,
                path: []const u8,
            },
            header: struct {
                key: []const u8,
                value: []const u8,
            },
            body: []const u8,
            // reached end of header
            end_of_header: void,
        };

        const Error = ParseError || ReaderType.Error;

        fn init(gpa: Allocator, buffer: []u8, reader: ReaderType) Self {
            return .{
                .gpa = gpa,
                .buffer = buffer,
                .reader = reader,
                .state = .status,
                .index = 0,
                .done = false,
                .content_length = 0,
                .header_start = 0,
                .header_end = 0,
                .chunked = false,
            };
        }

        fn nextEvent(self: *Self) Error!?Event {
            if (self.done) return null;

            return switch (self.state) {
                .status => self.parseStatus(),
                .header => self.parseHeader(),
                .body => self.parseBody(),
                .end_of_header => unreachable,
            };
        }

        fn parseStatus(self: *Self) Error!?Event {
            self.state = .header;
            const line = (try self.reader.readUntilDelimiterOrEof(self.buffer, '\n')) orelse return ParseError.EndOfStream;
            self.index += line.len + 1;
            self.header_start = self.index;
            var it = mem.tokenize(u8, try assertLE(line), " ");

            const parsed_method = it.next() orelse return ParseError.InvalidMethod;
            const parsed_path = it.next() orelse return ParseError.InvalidUrl;
            _ = it.next() orelse return ParseError.InvalidProtocol;

            return Event{
                .status = .{
                    .method = parsed_method,
                    .path = parsed_path,
                },
            };
        }

        fn parseHeader(self: *Self) Error!?Event {
            const line = (try self.reader.readUntilDelimiterOrEof(self.buffer[self.index..], '\n')) orelse return ParseError.EndOfStream;
            self.index += line.len + 1;
            if (line.len == 1 and line[0] == '\r') {
                self.header_end = self.index;
                if (self.content_length == 0 and !self.chunked) {
                    self.done = true;
                }
                self.state = .body;
                return Event.end_of_header;
            }
            var it = mem.split(u8, try assertLE(line), ": ");

            const key = it.next() orelse return ParseError.MissingHeaders;
            const value = it.next() orelse return ParseError.IncorrectHeader;

            // if content length hasn't been set yet,
            // check if it exists and set it by parsing the int value
            if (self.content_length == 0 and
                std.ascii.eqlIgnoreCase("content-length", key))
            {
                self.content_length = try std.fmt.parseInt(usize, value, 10);
            }

            // check if chunked body
            if (std.ascii.eqlIgnoreCase("transfer-encoding", key)) {
                // transfer-encoding can contain a list of encodings.
                // Therefore, iterate over them and check for 'chunked'.
                var split = std.mem.split(u8, value, ", ");
                while (split.next()) |maybe_chunk| {
                    if (std.ascii.eqlIgnoreCase("chunked", maybe_chunk)) {
                        self.chunked = true;
                    }
                }
            }

            return Event{
                .header = .{
                    .key = key,
                    .value = value,
                },
            };
        }

        fn parseBody(self: *Self) Error!?Event {
            defer self.done = true;

            if (self.content_length != 0) {
                const raw_body = try self.gpa.alloc(u8, self.content_length);
                try self.reader.readNoEof(raw_body);
                return Event{ .body = raw_body };
            }

            std.debug.assert(self.chunked);
            var body_list = std.ArrayList(u8).init(self.gpa);
            defer body_list.deinit();

            var read_len: usize = 0;
            while (true) {
                var len_buf: [1024]u8 = undefined; //Used to read the length of a chunk
                const lf_line = (try self.reader.readUntilDelimiterOrEof(&len_buf, '\n')) orelse
                    return error.InvalidBody;
                const line = try assertLE(lf_line);

                const index = std.mem.indexOfScalar(u8, line, ';') orelse
                    line.len;
                const chunk_len = try std.fmt.parseInt(usize, line[0..index], 10);
                try body_list.resize(read_len + chunk_len);
                try self.reader.readNoEof(body_list.items[read_len..]);
                read_len += chunk_len;

                // validate crlf
                var crlf: [2]u8 = undefined;
                try self.reader.readNoEof(&crlf);
                if (!std.mem.eql(u8, "\r\n", &crlf)) return error.InvalidBody;

                if (chunk_len == 0) {
                    break;
                }
            }
            return Event{ .body = try body_list.toOwnedSlice() };
        }

        fn assertLE(line: []const u8) ParseError![]const u8 {
            if (line.len == 0) return ParseError.InvalidLineEnding;
            const idx = line.len - 1;
            if (line[idx] != '\r') return ParseError.InvalidLineEnding;

            return line[0..idx];
        }
    };
}
