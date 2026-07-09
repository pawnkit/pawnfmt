enum
{
    STATE_NONE,
    Float:STATE_VALUES[4],
    STATE_LAST = 8,
};

enum Flags (<<= 1)
{
    FLAG_NONE = 1,
    FLAG_READ,
    FLAG_WRITE,
};

enum Offsets (= 5)
{
    OFFSET_X,
    OFFSET_Y,
};
