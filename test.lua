local ffi = require("ffi")

ffi.cdef [[
typedef struct Hold Hold;
int32_t handle_packets(struct Hold *hold, const char *bytes, unsigned int len);
struct Hold *init(void (*callback)(unsigned char*, unsigned int));
]]

local rust = ffi.load("package_handler")

local hold = rust.init(function(bytes, len)
    print("Received packet of length " .. len)
    print("First byte: " .. bytes[0])
end)

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
