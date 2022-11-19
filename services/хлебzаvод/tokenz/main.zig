const std = @import("std");
// const web = @import("./lib/web.zig");
const auth = @import("./src/auth.zig");
// const graceful = @import("./lib/graceful.zig");

// pub const io_mode = .evented;

pub fn main() !void {
    // var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    // var allocator = gpa.allocator();

    var svc = try auth.Service.init("auth-state");

    std.log.debug("{s} {}", .{ svc.secretKey, svc.nonce.value });

    // var server = try web.Server.init(gpa.allocator(), "0.0.0.0:80");

    // _ = async server.run();
    // try graceful.run();
}
