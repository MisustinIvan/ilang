extrn unit printf(string)

bool rule110(bool a, bool b, bool c) {
    if a && b && c { false };
    if a && b && !c { true };
}

int main() {
    let a : bool = rule110(true, true, false);
    printf("a = %d\n", a);
    1
}
