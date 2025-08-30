int main() {
    let a : int = 0;
    for a < 10 {
        a = a+1;
        printf("%d\n", a);
        let b : int = 0;
        for b < 10 {
            b = b+1;
            printf("%d:%d\n", a, b);
        };
    };
    return 1;
}
