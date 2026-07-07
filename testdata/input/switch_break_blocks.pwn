stock SwitchBreaks(value,flags)
{
	switch(value){
	case 0:{
		flags|=1;
		break;
	}
	case 1,2:
		if(flags)
			break;
	default:
		return flags;
	}
	return value+flags;
}