const std = @import("std");
const http = @import("http");
const os = std.os;
const auth = @import("./auth.zig");

pub const Server = struct {
    allocator: std.mem.Allocator,
    address: std.net.Address,
    context: Context,

    pub fn init(allocator: std.mem.Allocator, authSvc: *auth.Service, address: []const u8) !Server {
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

        return Server{
            .context = .{
                .authSvc = authSvc,
            },
            .allocator = allocator,
            .address = try std.net.Address.parseIp(host, port),
        };
    }

    pub fn run(self: *Server) void {
        self.runErr() catch |err| {
            std.log.err("error running http server: {}", .{err});
        };
    }

    fn runErr(self: *Server) !void {
        const builder = http.router.Builder(*Context);
        std.log.info("starting server on {}", .{self.address});

        try http.listenAndServe(self.allocator, self.address, &self.context, comptime http.router.Router(*Context, &.{
            builder.post("/sign", sign),
            builder.get("/unsign/:token", unsign),
        }));
    }
};

const Context = struct {
    authSvc: *auth.Service,
};

fn sign(ctx: *Context, response: *http.Response, request: http.Request) !void {
    if (request.body().len == 0) {
        return response.writeHeader(.bad_request);
    }

    const signed = try ctx.authSvc.sign(request.arena, request.body());
    try response.writer().writeAll(signed);
}

fn unsign(ctx: *Context, response: *http.Response, request: http.Request, token: []const u8) !void {
    const unsigned = ctx.authSvc.unsign(request.arena, token) catch |err| {
        switch (err) {
            error.InvalidToken => return response.writeHeader(.bad_request),
            else => return err,
        }
    };

    try response.headers.put("Content-Type", "application/octet-stream");
    try response.writer().writeAll(unsigned);
}
