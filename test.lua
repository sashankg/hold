local ffi = require("ffi")
local jit = require("jit")

jit.off()

ffi.cdef [[
typedef void (*callback_t)(const char* bytes, size_t len, const char* data);
typedef struct Hold Hold;
int32_t handle_packets(struct Hold *hold, const char *bytes, unsigned int len);
struct Hold *init_hold(void (*callback)(unsigned char*, unsigned int, void*), void*);
]]

local rust = ffi.load("package_handler")

local function callback(bytes, len, data)
    -- print("Received packet of length " .. len)
    -- print("First byte: " .. string.byte(bytes, 1))
    print(bytesToHex(bytes, len))
end

local cb = ffi.cast("callback_t", callback)

local hold = rust.init_hold(cb, ffi.cast("void*", "hello world"))

function bytesToHex(bytes, n)
    local hex = ""
    for i = 1, n do
        hex = hex .. string.format("%02X", bytes[i])
    end
    return hex
end

function hexToBytes(hex)
    local bytes = {}
    for i = 1, #hex, 2 do
        local byte = hex:sub(i, i + 1)
        table.insert(bytes, tonumber(byte, 16))
    end
    return bytes
end

local packet = hexToBytes(
    "45000040000040004006B1B20A0000057F000001EA2E1F900B47F39B00000000B002FFFF17B40000020405B4010303060101080A0418FE830000000004020000")

local cByteArray = ffi.new("unsigned char[?]", #packet, packet)

print(rust.handle_packets(hold, cByteArray, #packet))

while true do
end
