#if defined FEATURE_A
stock FeatureA()
{
    return 1;
}
#elseif defined(FEATURE_B)
new gFeatureValue = 2;
#else
enum FeatureMode
{
    FEATURE_NONE,
    FEATURE_FALLBACK = 3,
};
#endif
