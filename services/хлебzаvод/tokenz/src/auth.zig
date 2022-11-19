const std = @import("std");
const aes = std.crypto.aead.aes_gcm.Aes256Gcm;

pub const Service = struct {
    const tokenKeyLength = 4;

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

    fn generateTokenKey(_: *Service) ![Service.tokenKeyLength]u8 {
        var key: [Service.tokenKeyLength]u8 = undefined;
        std.crypto.random.bytes(key[0..]);

        return key;
    }

    fn generateNonce(_: *Service) ![aes.nonce_length]u8 {
        var nonce: [aes.nonce_length]u8 = undefined;
        std.crypto.random.bytes(nonce[0..]);

        return nonce;
    }

    fn encryptData(_: *Service, allocator: std.mem.Allocator, tokenKey: [Service.tokenKeyLength]u8, data: []const u8) ![]const u8 {
        const encrypted = try allocator.alloc(u8, data.len);
        std.mem.copy(u8, encrypted, data);

        for (data) |_, index| {
            encrypted[index] ^= tokenKey[index % tokenKey.len];
        }
        return encrypted;
    }
};
