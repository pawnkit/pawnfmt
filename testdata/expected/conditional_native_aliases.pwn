#if defined LEGACY_API
    native Float: GetActorHealth(actorid) = GetHealth;
    native SetActorSkin(actorid, skin) = AssignSkin;
#else
    forward Float: GetActorHealth(actorid);
#endif
