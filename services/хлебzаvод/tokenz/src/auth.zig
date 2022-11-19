const std = @import("std");
const aes = std.crypto.aead.aes_gcm.Aes256Gcm;

pub const Service = struct {
    random: DevUrandom,
    path: []const u8,
    secretKey: [aes.key_length]u8 = std.mem.zeroes([aes.key_length]u8),
    nonce: std.atomic.Atomic(u64) = .{ .value = 0 },

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

        var nonceBytes: [8]u8 = undefined;
        try stream.readNoEof(nonceBytes[0..]);
        self.nonce.store(std.mem.readIntSliceNative(u64, nonceBytes[0..]), .SeqCst);
    }

    fn create(self: *Service) !void {
        try self.random.read(self.secretKey[0..]);
        try self.save();
    }

    fn save(self: *Service) !void {
        var file = try std.fs.cwd().createFile(self.path, .{});
        defer file.close();

        var state: [aes.key_length + 8]u8 = undefined;
        std.mem.copy(u8, state[0..], self.secretKey[0..]);
        std.mem.writeIntSliceNative(u64, state[self.secretKey.len..], self.nonce.load(.SeqCst));

        try file.writeAll(state[0..]);
    }
};

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
