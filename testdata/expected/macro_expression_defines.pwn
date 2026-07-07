#define PICK(%0, %1, %2) %0 ? %1 : %2
#define SHIFT_MASK(%0, %1) (%0 << 1) & %1
#define SUB_AT(%0, %1) %0[%1]
#define CALL_WITH_SIZE(%0, %1) %0(%1, sizeof(%1))
