forward OnThing(playerid)<default>;
forward OnThing(playerid)<default,connected>;
stock OnThing(playerid)<default>{return 1;}

stock OnThing(playerid)<connected,idle>{return 1;}

stock static Float:GetScale(){return Float:(1.0);}