#assert MAX_PLAYERS>0&&defined LEGACY_MODE
#if !defined(FEATURE)||(MAX_PLAYERS<<1)>=64
#elseif defined(ALT_MODE)&&MAX_PLAYERS>16
#endif