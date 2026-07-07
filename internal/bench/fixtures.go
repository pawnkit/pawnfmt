package bench

import "strings"

const functionUnit = `
stock ProcessPlayer%d(playerid, Float:x, Float:y, Float:z)
{
    new Float:distance = floatsqroot((x * x) + (y * y) + (z * z));
    if (distance > 100.0)
    {
        SendClientMessage(playerid, 0xFFFFFFFF, "Too far away");
        return 0;
    }
    else if (distance > 50.0)
    {
        for (new i = 0; i < MAX_PLAYERS; i++)
        {
            if (!IsPlayerConnected(i)) continue;
            SendClientMessageToPlayer(i, playerid, distance);
        }
    }
    new arr[4] = {1, 2, 3, 4};
    switch (playerid % 3)
    {
        case 0: return arr[0];
        case 1, 2: return arr[1] + arr[2];
        default: return arr[3];
    }
}
`

func GenerateSource(n int) []byte {
	var b strings.Builder
	b.WriteString("#include <a_samp>\n#include <core>\n\n")
	for i := range n {
		b.WriteString(strings.ReplaceAll(functionUnit, "%d", itoa(i)))
	}
	return []byte(b.String())
}

func GenerateMacroHeavy(n int) []byte {
	var b strings.Builder
	for i := range n {
		s := itoa(i)
		b.WriteString("#define MAX_VALUE_" + s + " " + s + "\n")
		b.WriteString("#define SEND_MSG_" + s + "(%0,%1) SendClientMessage(%0,-1,%1)\n")
		b.WriteString("#define CLAMP_" + s + "(%0) (%0 < 0 ? 0 : (%0 > 100 ? 100 : %0))\n")
	}
	return []byte(b.String())
}

func GeneratePreprocessorHeavy(n int) []byte {
	var b strings.Builder
	for i := range n {
		s := itoa(i)
		b.WriteString("#if defined FEATURE_" + s + "\n")
		b.WriteString("new gFeatureFlag" + s + " = 1;\n")
		b.WriteString("#else\n")
		b.WriteString("new gFeatureFlag" + s + " = 0;\n")
		b.WriteString("#endif\n\n")
	}
	return []byte(b.String())
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
