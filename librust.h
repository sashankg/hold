#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

typedef struct Hold Hold;

int32_t handle_packets(struct Hold *hold, char *bytes, unsigned int len);

struct Hold *init(void (*callback)(unsigned char*, unsigned int));
