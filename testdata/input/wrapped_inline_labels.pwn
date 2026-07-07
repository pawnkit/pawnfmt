stock Wrapped(value)
{
#if defined FEATURE
label_a: return 1;
#elseif defined ALT
    while (value > 0) {
    #if defined INNER
        retry: continue;
    #endif
    }
#else
    guard: #warning fallback branch
#endif
}