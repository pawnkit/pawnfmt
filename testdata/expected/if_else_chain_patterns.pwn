stock Chain(value, a, b)
{
    if (value < 0)
    {
        return -1;
    }
    else if (value == 0)
    {
        return 0;
    }
    else if (a > b)
    {
        value = a - b;
    }
    else
    {
        value = a + b;
    }

    if (value > 10)
    {
        if (a)
        {
            return value;
        }
        else
        {
            return b;
        }
    }
    else if (value > 5)
    {
        return a;
    }

    return value;
}
