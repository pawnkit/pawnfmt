stock InlineLabelDirectives(value)
{
    while (value > 0)
    {
        next_pass: continue;
    }

    emit_here: #emit ABS
#if defined TESTING
    guard_here: #warning debug build
#endif
}
