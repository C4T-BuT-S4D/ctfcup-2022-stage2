const std = @import("std");
// const web = @import("./lib/web.zig");
const auth = @import("./src/auth.zig");
const graceful = @import("./src/graceful.zig");

pub const io_mode = .evented;

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    var allocator = gpa.allocator();

    var svc = try auth.Service.init(".secret-key");

    var token = try svc.sign(allocator, "aboba");
    std.log.debug("{x}", .{std.fmt.fmtSliceHexLower(svc.secretKey[0..])});
    std.log.debug("{s}", .{token});
    std.log.debug("{x}", .{std.fmt.fmtSliceHexLower(try svc.unsign(allocator, token))});

    // var server = try web.Server.init(gpa.allocator(), "0.0.0.0:80");

    // _ = async server.run();
    try graceful.run(&svc);
}
