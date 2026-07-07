enum Mode {
#if defined FEATURE_A
    MODE_A,
    MODE_B,
#else
    MODE_C,
#endif
    MODE_DONE
};