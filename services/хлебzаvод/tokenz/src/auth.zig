const std = @import("std");
const aes = std.crypto.aead.aes_gcm.Aes256Gcm;

pub const Service = struct {
    const tokenKeyLength = 16;

    random: DevUrandom,
    path: []const u8,
    secretKey: [aes.key_length]u8 = std.mem.zeroes([aes.key_length]u8),

    pub fn init(path: []const u8) !Service {
        var service = Service{
            .random = try DevUrandom.init(),
            .path = path,
        };

        try service.load(path);
        return service;
    }

    fn load(self: *Service, path: []const u8) !void {
        var file = std.fs.cwd().openFile(path, .{}) catch |err| {
            if (err == std.fs.File.OpenError.FileNotFound) {
                try self.create();
                return;
            }
            return err;
        };
        defer file.close();

        const stream = file.reader();
        try stream.readNoEof(self.secretKey[0..]);
    }

    fn create(self: *Service) !void {
        try self.random.read(self.secretKey[0..]);
        try self.save();
    }

    pub fn save(self: *Service) !void {
        var file = try std.fs.cwd().createFile(self.path, .{});
        defer file.close();

        try file.writeAll(self.secretKey[0..]);
    }

    pub fn sign(self: *Service, allocator: std.mem.Allocator, data: []const u8) ![]const u8 {
        const tokenKey = try self.generateTokenKey();
        const encryptedData = try self.encryptData(allocator, tokenKey, data);

        var nonce = try self.generateNonce();
        var encryptedTokenKey: [Service.tokenKeyLength]u8 = undefined;
        var tag: [aes.tag_length]u8 = undefined;

        aes.encrypt(encryptedTokenKey[0..], &tag, tokenKey[0..], encryptedData, nonce, self.secretKey);

        const encoder = std.base64.url_safe_no_pad.Encoder;
        const parts = [_][]const u8{ encryptedData, encryptedTokenKey[0..], nonce[0..], tag[0..] };
        var size: usize = 0;
        for (parts) |part, index| {
            size += encoder.calcSize(part.len);
            if (index < parts.len - 1) {
                size += 1;
            }
        }

        const token = try allocator.alloc(u8, size);
        var position: usize = 0;
        for (parts) |part, index| {
            position += encoder.encode(token[position..], part).len;
            if (index < parts.len - 1) {
                token[position] = '.';
                position += 1;
            }
        }

        return token;
    }

    pub fn unsign(self: *Service, allocator: std.mem.Allocator, data: []const u8) ![]const u8 {
        if (std.mem.count(u8, data, ".") != 3) {
            return error.InvalidToken;
        }

        const decoder = std.base64.url_safe_no_pad.Decoder;
        var spliterator = std.mem.split(u8, data, ".");
        var parts: [4][]const u8 = undefined;
        for (parts) |_, index| {
            parts[index] = spliterator.next() orelse unreachable;

            var decoded = try allocator.alloc(
                u8,
                decoder.calcSizeForSlice(parts[index]) catch return error.InvalidToken,
            );
            decoder.decode(decoded, parts[index]) catch return error.InvalidToken;
            parts[index] = decoded;
        }

        var encryptedData = parts[0];
        var encryptedTokenKey = parts[1];
        var nonce = parts[2];
        var tag = parts[3];

        if (encryptedTokenKey.len != Service.tokenKeyLength) {
            return error.InvalidToken;
        } else if (nonce.len != aes.nonce_length) {
            return error.InvalidToken;
        } else if (tag.len != aes.tag_length) {
            return error.InvalidToken;
        }

        var tokenKey = try allocator.alloc(u8, Service.tokenKeyLength);
        aes.decrypt(
            tokenKey,
            encryptedTokenKey,
            tag[0..aes.tag_length].*,
            encryptedData,
            nonce[0..aes.nonce_length].*,
            self.secretKey,
        ) catch return error.InvalidToken;

        return tokenKey;
    }

    fn generateTokenKey(self: *Service) ![Service.tokenKeyLength]u8 {
        var key: [Service.tokenKeyLength]u8 = undefined;
        try self.random.read(key[0..]);

        return key;
    }

    fn generateNonce(self: *Service) ![aes.nonce_length]u8 {
        var nonce: [aes.nonce_length]u8 = undefined;
        try self.random.read(nonce[0..]);

        return nonce;
    }

    fn encryptData(_: *Service, allocator: std.mem.Allocator, tokenKey: [Service.tokenKeyLength]u8, data: []const u8) ![]const u8 {
        const N: usize = 256;
        var state: [N]u8 = undefined;
        var i: u8 = 0;
        while (true) : (i += 1) {
            state[i] = i;

            if (i == N - 1) {
                break;
            }
        }

        var j: u8 = 0;
        i = 0;
        while (i < N) : (i += 1) {
            var t: u8 = undefined;
            _ = @addWithOverflow(u8, state[i], tokenKey[i % tokenKey.len], &t);
            _ = @addWithOverflow(u8, j, t, &j);

            std.mem.swap(u8, &state[i], &state[j]);
            if (i == N - 1) {
                break;
            }
        }

        const encrypted = try allocator.alloc(u8, data.len);
        std.mem.copy(u8, encrypted, data);

        i = 0;
        j = 0;
        for (data) |_, index| {
            _ = @addWithOverflow(u8, i, 1, &i);
            _ = @addWithOverflow(u8, j, state[i], &j);

            std.mem.swap(u8, &state[i], &state[j]);

            var t: u8 = undefined;
            _ = @addWithOverflow(u8, state[i], state[j], &t);

            encrypted[index] ^= state[t];
        }
        return encrypted;
    }
};

// Zig is so shit that std.crypto.rand doesn't work on Linux with io_mode = .evented
const DevUrandom = struct {
    fd: std.os.fd_t,
    file: std.fs.File,
    br: std.io.BufferedReader(4096, std.fs.File.Reader),
    mu: std.Thread.Mutex,

    fn init() !DevUrandom {
        const fd = try std.os.openZ("/dev/urandom", std.os.system.O.RDONLY | std.os.system.O.CLOEXEC, 0);
        const file = std.fs.File{
            .handle = fd,
            .capable_io_mode = .blocking,
            .intended_io_mode = .blocking,
        };

        return DevUrandom{
            .fd = fd,
            .file = file,
            .br = std.io.bufferedReader(file.reader()),
            .mu = .{},
        };
    }

    fn read(self: *DevUrandom, buf: []u8) !void {
        self.mu.lock();
        defer self.mu.unlock();

        try self.br.reader().readNoEof(buf);
    }
};
