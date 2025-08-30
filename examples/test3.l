int main() {
    let a : int = rule110(3);
    printf("a = %d\n", a);
    1
}

int rule110(int a, int b, int c) {
    if a && b && c { 0 };
    if a && b && !c { 0 };
}
