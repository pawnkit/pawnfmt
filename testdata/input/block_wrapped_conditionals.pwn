stock ConfigureFeature(playerid)
{
#if defined FEATURE_A
    new enabled = 1;
#elseif defined(FEATURE_B)
    if (playerid > 0) {
        return 2;
    }
#else
    return 0;
#endif

    return 1;
}