const std = @import("std");
const aes = std.crypto.aead.aes_gcm.Aes256Gcm;

pub const Service = struct {
    path: []const u8,
    secretKey: [aes.key_length]u8 = std.mem.zeroes([aes.key_length]u8),

    pub fn init(path: []const u8) !Service {
        var service = Service{
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
        std.crypto.random.bytes(self.secretKey[0..]);
        try self.save();
    }

    pub fn save(self: *Service) !void {
        var file = try std.fs.cwd().createFile(self.path, .{});
        defer file.close();

        try file.writeAll(self.secretKey[0..]);
    }

    pub fn sign(self: *Service, allocator: std.mem.Allocator, data: []const u8) ![]const u8 {
        var encryptedData = try allocator.alloc(u8, data.len);
        var nonce = try self.generateNonce();
        var tag: [aes.tag_length]u8 = undefined;

        aes.encrypt(encryptedData, &tag, data, &.{}, nonce, self.secretKey);

        const encoder = std.base64.url_safe_no_pad.Encoder;
        const parts = [_][]const u8{ encryptedData, nonce[0..], tag[0..] };
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
        if (std.mem.count(u8, data, ".") != 2) {
            return error.InvalidToken;
        }

        const decoder = std.base64.url_safe_no_pad.Decoder;
        var spliterator = std.mem.split(u8, data, ".");
        var parts: [3][]const u8 = undefined;
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
        var nonce = parts[1];
        var tag = parts[2];

        if (nonce.len != aes.nonce_length) {
            return error.InvalidToken;
        } else if (tag.len != aes.tag_length) {
            return error.InvalidToken;
        }

        var decryptedData = try allocator.alloc(u8, encryptedData.len);
        aes.decrypt(
            decryptedData,
            encryptedData,
            tag[0..aes.tag_length].*,
            &.{},
            nonce[0..aes.nonce_length].*,
            self.secretKey,
        ) catch return error.InvalidToken;

        return decryptedData;
    }

    fn generateNonce(_: *Service) ![aes.nonce_length]u8 {
        var nonce: [aes.nonce_length]u8 = undefined;
        std.crypto.random.bytes(nonce[0..]);

        return nonce;
    }
};
