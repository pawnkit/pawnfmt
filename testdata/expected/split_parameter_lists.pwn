stock Foo(
#if defined FEATURE_A
    playerid,
#else
    playerid,
    value,
#endif
    tail
)
{
    return 1;
}
