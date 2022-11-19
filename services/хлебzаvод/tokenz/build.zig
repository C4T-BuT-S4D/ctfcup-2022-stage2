const std = @import("std");

pub fn build(b: *std.build.Builder) void {
    const target = b.standardTargetOptions(.{});
    const mode = b.standardReleaseOptions();

    const exe = b.addExecutable("app", "main.zig");
    exe.setTarget(target);
    exe.setBuildMode(mode);
    exe.use_stage1 = true;

    exe.addPackage(.{ .name = "http", .source = .{ .path = "src/http/http.zig" } });

    exe.install();
}
