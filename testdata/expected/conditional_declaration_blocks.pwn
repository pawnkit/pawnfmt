#if defined DEBUG
forward DebugLog(const message[]);

new gDebugLevel = 2;
#else
native WriteLog(const message[]);
#endif
