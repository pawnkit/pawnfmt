stock Looping(value)
{
    for (new i = 0; i < 10; i++)
    {
        if (i == 5)
        {
            continue;
        }
    }
    while (value > 0)
    {
        value--;
    }
    do
    {
        value--;
    }
    while (value > 0);
}
