stock Weird(x, y)
{
    new z = x+ /* mid-expr, kept raw */ y;
    return z;
}

enum Mode
{
    A, // a comment inside an enum body
    B,
};

stock CallWithComment()
{
    Foo(1, /* stray */ 2);
    return 1;
}

stock Clean(a, b)
{
    return a + b;
}
