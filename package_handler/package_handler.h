#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

typedef struct Hold Hold;

int32_t handle_packets(struct Hold *hold, const unsigned char *bytes, size_t len);

struct Hold *init_hold(void);
