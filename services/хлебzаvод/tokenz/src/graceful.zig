const std = @import("std");
const os = std.os;

const shutdownDelayMs = 500;
var shutdownMutex = std.Thread.Mutex{};
var shutdownCondition = std.Thread.Condition{};

pub fn run(auth: anytype) !void {
    var gracefulShutdownAct = os.Sigaction{
        .handler = .{ .handler = shutdown },
        .mask = os.empty_sigset,
        .flags = 0,
    };
    try os.sigaction(os.SIG.INT, &gracefulShutdownAct, null);
    try os.sigaction(os.SIG.TERM, &gracefulShutdownAct, null);

    shutdownMutex.lock();
    defer shutdownMutex.unlock();
    shutdownCondition.wait(&shutdownMutex);

    auth.save() catch |err| {
        std.log.err("failed to save auth state: {s}", .{@errorName(err)});
    };

    std.log.info("shutting down in {}ms", .{shutdownDelayMs});
    std.time.sleep(shutdownDelayMs * std.time.ns_per_ms);
    std.os.exit(0);
}

pub fn shutdown(_: c_int) align(1) callconv(.C) void {
    shutdownMutex.lock();
    defer shutdownMutex.unlock();

    shutdownCondition.signal();
}
