stock InlineLabelBranches(value)
{
    enter: {
        return 1;
    }
    countdown: for (; value > 0; value--)
    {
        break;
    }
    retry: do
    {
        value--;
    }
    while (value > 0);
}
