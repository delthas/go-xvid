#ifndef GOXVID_H
#define GOXVID_H

#include <stdbool.h>
#include <stdlib.h>
#include <xvid.h>

// go throws overflow errors when using those directly (value = 1<<31)
// redeclare them with an explicit type
extern const unsigned int CPU_FORCE;
extern const unsigned int DEBUG_DEBUG;
extern const unsigned int CSP_VFLIP;

typedef struct {
  int general;
  int time_base;
  int time_increment;
  int * qscale;
  int qscale_stride;
} vop_t;

vop_t vop_data(xvid_dec_stats_t *stats);

typedef struct {
  int general;
  int width;
  int height;
  int par;
  int par_width;
  int par_height;
} vol_t;

vol_t vol_data(xvid_dec_stats_t *stats);

int pluginCallback_cgo(void * handle, int opt, void * param1, void * param2);

#endif //GOXVID_H
