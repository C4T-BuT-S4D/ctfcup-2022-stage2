const std = @import("std");
const http = @import("http");
const os = std.os;

pub const Server = struct {
    allocator: std.mem.Allocator,
    address: std.net.Address,
    context: Context,

    pub fn init(allocator: std.mem.Allocator, address: []const u8) !Server {
        var it = std.mem.split(u8, address, ":");
        var host: []const u8 = undefined;
        var port: u16 = undefined;
        if (it.next()) |part| {
            host = part;
        }
        if (it.next()) |part| {
            port = try std.fmt.parseUnsigned(u16, part, 0);
        }
        if (it.rest().len > 0) {
            return error.InvalidIPAddressFormat;
        }

        return Server{ .context = .{}, .allocator = allocator, .address = try std.net.Address.parseIp(host, port) };
    }

    pub fn run(self: *Server) void {
        self.runErr() catch |err| {
            std.log.err("running http server: {}", .{err});
        };
    }

    fn runErr(self: *Server) !void {
        const builder = http.router.Builder(*Context);

        try http.listenAndServe(self.allocator, self.address, &self.context, comptime http.router.Router(*Context, &.{
            builder.get("/sign", sign),
            builder.get("/unsign/:token"),
        }));
    }
};

const Context = struct {};

fn sign(_: *Context, response: *http.Response, request: http.Request) !void {}

fn unsign(_: *Context, response: *http.Response, request: http.Request) !void {}

fn index(_: *Context, response: *http.Response, _: http.Request) !void {
    try response.headers.put("Content-Type", "text/html");
    try response.writer().writeAll(
        \\<!DOCTYPE html>
        \\<html>
        \\<head>
        \\    <meta charset="utf-8" />
        \\</head>
        \\<body>
        \\    <a href="/metrics">Metrics</a>
        \\</body>
        \\</html>
        \\
    );
}

fn metrics(_: *Context, response: *http.Response, _: http.Request) !void {
    const flag = os.getenv("FLAG") orelse "fakeflag";

    try response.headers.put("Content-Type", "text/plain; version=0.0.4");
    try response.writer().print(
        \\app{{flag="{s}"}} 1
        \\
    , .{flag});
}
