#define TEMP_VALUE 1
#define TEMP_MACRO(%0) (%0)
#undef TEMP_VALUE
#if defined TEMP_MACRO
#undef TEMP_MACRO
#endif
