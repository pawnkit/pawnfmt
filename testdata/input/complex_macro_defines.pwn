#define DIALOG_TITLE "Server"
#define SEND_MSG(%0,%1) SendClientMessage(%0,0xFFFFFFFF,%1)
#define SET_ELEM(%0,%1,%2) %0[%1]=%2
#define WITH_SIZE(%0) sizeof(%0)
#define MULTI_STMT(%0,%1) { %0 = %1; %1 = %0; }
#define IF_ELSE_WRAP(%0) if (%0) return 1; else return 0
#define DO_WRAP(%0,%1) do { %0; } while (%1)
#define SWITCH_MULTI(%0) switch (%0) { case 1,3 .. 5: return 9; default: break; }