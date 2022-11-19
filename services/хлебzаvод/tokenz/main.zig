const std = @import("std");
const web = @import("./src/web.zig");
const auth = @import("./src/auth.zig");
const graceful = @import("./src/graceful.zig");

pub const io_mode = .evented;

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    var allocator = gpa.allocator();

    var svc = try auth.Service.init(".secret-key");
    var server = try web.Server.init(allocator, &svc, "0.0.0.0:80");

    _ = async server.run();
    try graceful.run(&svc);
}
