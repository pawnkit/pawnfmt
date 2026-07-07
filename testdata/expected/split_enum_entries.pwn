enum Mode
{
#if defined FEATURE_A
    MODE_A,
#else
    MODE_B,
#endif
    MODE_C,
};
