#include "goxvid.h"

const unsigned int CPU_FORCE = XVID_CPU_FORCE;
const unsigned int DEBUG_DEBUG = XVID_DEBUG_DEBUG;
const unsigned int CSP_VFLIP = XVID_CSP_VFLIP;

vop_t vop_data(xvid_dec_stats_t *stats) {
  return (vop_t) { .general = stats->data.vop.general, .time_base = stats->data.vop.time_base, .time_increment = stats->data.vop.time_increment, .qscale = stats->data.vop.qscale, .qscale_stride = stats->data.vop.qscale_stride };
}

vol_t vol_data(xvid_dec_stats_t *stats) {
  return (vol_t) { .general = stats->data.vol.general, .width = stats->data.vol.width, .height = stats->data.vol.height, .par = stats->data.vol.par, .par_width = stats->data.vol.par_width, .par_height = stats->data.vol.par_height };
}

int pluginCallback_cgo(void * handle, int opt, void * param1, void * param2) {
  int pluginCallback(void*, int, void*, void*);
  return pluginCallback(handle, opt, param1, param2);
}
