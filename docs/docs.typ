#set text(font: "Iosevka NF")
#set page(numbering: "1.")
#set heading(numbering: "1.")
#set raw(syntaxes: ("syntax.sublime-syntax",))

#import "@preview/nutthead-ebnf:0.3.1": *

#align(center)[
  #text(size: 20pt)[
    *Dokumentace jazyka Ilang*
  ]
]

#v(8em)

#outline(title: [Obsah])
#pagebreak()

= Gramatika


Níže je vypsaná formální gramatika jazyka ve formě ekvivalentí s rozšířenou Backus-Naurovou formou.

#context [
  #ebnf[
    #[

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [program],
        definition-list: (
          [
            #repeated-sequence(
              [#grouped-sequence(
                [#single-definition[external_declaration]],
                [#single-definition[function_declaration]],
                [#single-definition[comment]],
            )])
          ],
        )
      )
]


#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [comment],
        definition-list: (
          [
            #terminal(illumination: "highlighted")[\#]
            #special-sequence[any characters]
            #terminal(illumination: "highlighted")[\\n]
          ],
        )
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [external_declaration],
        definition-list: (
          (indent: 1),
          [
            #terminal(illumination: "highlighted")[extrn]
            #single-definition[type]
            #single-definition[identifier]
            #terminal(illumination: "highlighted")[(]
          ],
          (indent: 2),
          [
            #grouped-sequence([
              #optional-sequence([
                #single-definition([argument])
                #repeated-sequence([
                  #terminal(illumination: "highlighted")[,]
                  #single-definition[argument]
                ])
                #optional-sequence([
                  #terminal(illumination: "highlighted")[,]
                  #terminal(illumination: "highlighted")[...]
                ])
              ],)
            ],
            [
              #terminal(illumination: "highlighted")[...]
            ],)
          ],
          (indent: 1),
          [
            #terminal(illumination: "highlighted")[)]
          ],
        )
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [function_declaration],
        definition-list: (
          (indent: 1),
          [
            #single-definition[type]
            #single-definition[identifier]
            #terminal(illumination: "highlighted")[(]
          ],
          (indent: 2),
          [
            #optional-sequence([
              #single-definition([argument])
              #repeated-sequence([
                #terminal(illumination: "highlighted")[,]
                #single-definition[argument]
              ])
            ],)
          ],
          (indent: 1),
          [
            #terminal(illumination: "highlighted")[)]
            #single-definition[block_expression]
          ],
        )
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [argument],
        definition-list: ([
          #single-definition[type]
          #single-definition[identifier]
        ],) 
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [block_expression],
        definition-list: ([
          #terminal(illumination: "highlighted")[{]
          #repeated-sequence([
            #single-definition[expression]
            #terminal(illumination: "highlighted")[;]
          ],)
          #optional-sequence([
            #single-definition[expression]
          ],)
          #terminal(illumination: "highlighted")[}]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [expression],
        definition-list: ([
          #grouped-sequence(
          [#single-definition[return]],
          [#single-definition[bind]],
          [#single-definition[assignment]],
          [#single-definition[value]],
          )
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [value],
        definition-list: ([
          #grouped-sequence(
            [#single-definition[primary]],
            [#single-definition[binary]],
            [#single-definition[unary]],
          )
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [return],
        definition-list: ([
          #terminal(illumination: "highlighted")[return]
          #single-definition[value]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [bind],
        definition-list: ([
          #terminal(illumination: "highlighted")[let]
          #single-definition[identifier]
          #terminal(illumination: "highlighted")[:]
          #single-definition[type]
          #terminal(illumination: "highlighted")[=]
          #single-definition[value]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [assignment],
        definition-list: ([
          #grouped-sequence(
            [#single-definition[identifier]],
            [#single-definition[index]],
            [#single-definition[deref]],
          )
          #terminal(illumination: "highlighted")[=]
          #single-definition[value]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [deref],
        definition-list: ([
          #terminal(illumination: "highlighted")[@]
          #single-definition[identifier]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [binary],
        definition-list: ([
          #single-definition[primary]
          #single-definition[binary_operator]
          #single-definition[value]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [unary],
        definition-list: ([
          #single-definition[unary_operator]
          #single-definition[primary]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [index],
        definition-list: ([
          #single-definition[identifier]
          #terminal(illumination: "highlighted")[\[]
          #single-definition[value]
          #terminal(illumination: "highlighted")[\]]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [primary],
        definition-list: (
          grouped-sequence(
          [#single-definition[literal]],
          [#single-definition[identifier]],
          [#single-definition[call]],
          [#single-definition[separated]],
          [#single-definition[block]],
          [#single-definition[condition]],
          [#single-definition[index]],
          [#single-definition[deref]],
          [#single-definition[loop]],
          [#single-definition[make]],
          [#single-definition[release]],
          ),
        )
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [make],
        definition-list: ([
          #terminal(illumination: "highlighted")[make]
          #terminal(illumination: "highlighted")[(]
          #single-definition[basic_type]
          #terminal(illumination: "highlighted")[,]
          #single-definition[value]
          #terminal(illumination: "highlighted")[)]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [release],
        definition-list: ([
          #terminal(illumination: "highlighted")[release]
          #terminal(illumination: "highlighted")[(]
          #single-definition[identifier]
          #terminal(illumination: "highlighted")[)]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [loop],
        definition-list: ([
          #terminal(illumination: "highlighted")[for]
          #single-definition[value]
          #single-definition[block_expression]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [literal],
        definition-list: ([
          #grouped-sequence(
            [#single-definition[basic_literal]],
            [#single-definition[array_literal]],
          )
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [basic_literal],
        definition-list: ([
          #grouped-sequence(
            [#single-definition[int_literal]],
            [#single-definition[float_literal]],
            [#single-definition[string_literal]],
            [#single-definition[bool_literal]],
          )
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [array_literal],
        definition-list: ([
          #terminal(illumination: "highlighted")[\[]
          #single-definition[value]
          #repeated-sequence([
            #terminal(illumination: "highlighted")[,]
            #single-definition[value]
          ],)
          #terminal(illumination: "highlighted")[\]]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [call],
        definition-list: ([
          #single-definition[identifier]
          #terminal(illumination: "highlighted")[(]
          #single-definition[value]
          #repeated-sequence([
            #terminal(illumination: "highlighted")[,]
            #single-definition[value]
          ],)
          #terminal(illumination: "highlighted")[)]
        ],)
      )
]


#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [separated],
        definition-list: ([
          #terminal(illumination: "highlighted")[(]
          #single-definition[value]
          #terminal(illumination: "highlighted")[)]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [condition],
        definition-list: ([
          #terminal(illumination: "highlighted")[if]
          #single-definition[value]
          #single-definition[value]
          #optional-sequence([
            #terminal(illumination: "highlighted")[else]
            #single-definition[value]
          ],)
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [identifier],
        definition-list: ([
          #single-definition[letter]
          #repeated-sequence([
            #grouped-sequence(
              [#single-definition[letter]],
              [#single-definition[digit]],
              [#terminal(illumination: "highlighted")[\_]],
            )
          ],)
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [int_literal],
        definition-list: ([
          #single-definition[digit]
          #repeated-sequence(
              [#single-definition[digit]],
          )
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [float_literal],
        definition-list: ([
          #single-definition[digit]
          #repeated-sequence(
              [#single-definition[digit]],
          )
          #terminal(illumination: "highlighted")[.]
          #single-definition[digit]
          #repeated-sequence(
              [#single-definition[digit]],
          )
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [string_literal],
        definition-list: ([
          #terminal(illumination: "highlighted")["]
          #special-sequence[any characters]
          #terminal(illumination: "highlighted")["]
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [bool_literal],
        definition-list: ([
          #grouped-sequence(
            [#terminal(illumination: "highlighted")[true]],
            [#terminal(illumination: "highlighted")[false]],
          )
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [type],
        definition-list: ([
          #grouped-sequence(
            [#single-definition[basic_type]],
            [#single-definition[array_type]],
            [#single-definition[slice_type]],
            [#single-definition[pointer_type]],
          )
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [basic_type],
        definition-list: ([
          #grouped-sequence(
            [#terminal(illumination: "highlighted")[int]],
            [#terminal(illumination: "highlighted")[bool]],
            [#terminal(illumination: "highlighted")[float]],
            [#terminal(illumination: "highlighted")[string]],
            [#terminal(illumination: "highlighted")[unit]],
          )
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [pointer_type],
        definition-list: ([
          #terminal(illumination: "highlighted")[^] 
          #single-definition[basic_type] 
        ],)
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [binary_operator],
        definition-list: (
          grouped-sequence(
          [#terminal(illumination: "highlighted")[+]],
          [#terminal(illumination: "highlighted")[-]],
          [#terminal(illumination: "highlighted")[\*]],
          [#terminal(illumination: "highlighted")[/]],
          [#terminal(illumination: "highlighted")[==]],
          [#terminal(illumination: "highlighted")[!=]],
          [#terminal(illumination: "highlighted")[<]],
          [#terminal(illumination: "highlighted")[>]],
          [#terminal(illumination: "highlighted")[<=]],
          [#terminal(illumination: "highlighted")[>=]],
          [#terminal(illumination: "highlighted")[<<]],
          [#terminal(illumination: "highlighted")[>>]],
          [#terminal(illumination: "highlighted")[||]],
          ),
        )
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [unary_operator],
        definition-list: (
          grouped-sequence(
          [#terminal(illumination: "highlighted")[-]],
          [#terminal(illumination: "highlighted")[!]],
          [#terminal(illumination: "highlighted")[^]],
          [#terminal(illumination: "highlighted")[@]],
          ),
        )
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [letter],
        definition-list: (
          grouped-sequence(
          [#special-sequence[a ... z]],
          [#special-sequence[A ... Z]],
          ),
        )
      )
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
      #syntax-rule(
        meta-id: [digit],
        definition-list: (
          [#special-sequence[0 ... 9]],
        )
      )
]

    ]
  ]
]

#pagebreak()
= Sémantika
== Typy
Jazyk je staticky typovaný. Každý výraz má typ určený při překladu.

Základní typy jsou:
- *int* - 64bitové celé číslo se znaménkem
- *float* - 64bitové desetinné číslo, ekvivalent *double* v C
- *bool* - pravdivostní hodnota *true*(1) nebo *false*(0)
- *string* - odkaz na řetězec znaků zakončený nulovým znakem
- *unit* - typ bez velikosti, používán pro výrazy bez návratové hodnoty

Složené typy jsou:
- *[N]T* - pole *N* prvků typu *T* alokované na zásobníku kde *N* je přirozené kladné číslo známé při překladu
- *[R]T* - odkaz na pole *R* prvků typu *T* alokované na zásobníku nebo haldě
- *[]T* - odkaz na pole prvků typu *T* alokované na zásobníku nebo haldě
- *^T* - odkaz na hodnotu základního typu *T*

== Rozsah platnosti
Jazyk má pro proměnné jediný rozsah platnosti, a to v těle funkce. Všechny proměnné alokované pomocí *let* jsou přístupné od místa deklarace do konce těla funkce. Externí funkce deklarované pomocí *extrn* jsou dostupné v těle každé funkce nezávisle na pořadí deklarace. Normální funkce jsou dostupné ve vlastním těle(pro podporu rekurze) a v těle všech následujících funkcí.

== Hodnoty a výrazy
Každý výraz vrací hodnotu. Bloky *{ ... }* vrací hodnotu posledního výrazu v těle, pokud za ním není středník *;*. V případě, že poslední výraz za sebou středník *;* má, vrací *0*. Podmínka *if* vrací hodnotu větve, která byla vyhodnocena. Cyklus *for* vrací poslední hodnotu těla, nebo *0* pokud tělo neproběhlo ani jednou.

== Správa paměti
Pole(arrays) jsou alokována na zásobníku a jejich velikost musí být známa při překladu. Dynamická pole(slices) jsou buďto odkazy na pole na zásobníku(tak jsou pole předávána do funkcí) nebo jsou alokována na haldě pomocí *make(T, N)*, kde *T* je základní typ a *N* je počet prvků, který nemusí být při překladu známý. Vestavěná funkce *make* je abstrakcí nad funkcí z libc *malloc*. Správa paměti je plně na uživateli, a tak musí alokovanou paměť dealokovat pomocí vestavěné funkce *relase(S)* kde *S* je identifikátor s typem *slice*. Vestavěná funkce *release* je abstrakcí nad funckcí z libc *free*.

== Předávání argumentů
Argumenty jsou předávány hodnotou. Pole a odkazy na pole jsou předávány jako dvojice (*odkaz* *délka*). Úpravy prvků pole uvnitř funkce se tedy projeví i mimo ni, přiřazení pole ale nikoliv. To jenom upraví hodnotu odkazu a délky v lokální proměnné. Výjimka je pro argumenty s typem pole, kde známe délku při překladu. V tom případě dojde při přiřazení k hodnotě se stejným typem k překopírování prvků.

== Precedence operátorů
Precedence operátorů není přímo definována gramatikou, ale následující tabulkou:

#grid(
  columns: 2,
  rows: 22pt,
  stroke: 0.5pt + black,
  inset: 5pt,
  align: horizon,
  [1], [||],
  [2], [&&],
  [3], [==, !=],
  [4], [<, >, <=, >=],
  [5], [+, -],
  [6], [\*, /],
  [7], [<<, >>],
)

== Omezení
Logické operátory nepodporují zkrácené vyhodnocení, vždy se vyhodnotí celý výraz.
Indexování je povoleno jenom pro hodnoty typu *slice* nebo *array*, ne pro hodnoty typu *string*.
Koerce typů je podporována jen pro implicitní převod hodnoty typu *array* na odkaz typu *slice*.

#pagebreak()
= Psaní programů
== Hello World
Jednoduchý program vypisující text na standardní výstup:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
extrn unit printf(string format, ...)

int main() {
  printf("Hello, world!");
  0
}
```
]

== Práce s různými typy
#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
let x: int = 42;
let pi: float = 3.14;
let b: bool = false;
let s: string = "text";
```
]

== Deklarace funkcí
Jednoduchá funkce vracející součet dvou argumentů:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
extrn unit printf(string format, ...)

int add2(int a, int b) {
  a+b
}

int main() {
  printf("Hello, world!");
  0
}
```
]

== Externí funkce
Ukázka použití funkcí z *\<math.h\>*, při překladu pomocí *gcc* potom nutno poskytnout linker flag *-lm*

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
extrn unit srand(int seed)
extrn int time(int loc)
extrn int rand()

srand(time(0));
let random_number: int = rand();
```
]

#pagebreak()

== Vstup a výstup
Pro vstup a výstup se používají funkce ze standardní knihovny C:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
extrn unit printf(string fmt, ...)
extrn int scanf(string fmt, ...)
extrn int putchar(int c)
extrn int getchar()
```
]

== Práce s poli
Statické pole fixní velikosti:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
let arr: [8]int = [1, 2, 3, 4, 5, 6, 7, 8];
arr[3] = 13;
let x: int = arr[3];  # x = 13
```
]

Dynamické pole alokované na haldě:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
let n: int = 100;
let arr: [arr_len]int = make(int, n);
arr[0] = 42;
release(arr);
```
]

Předání pole do funkce a přiřazení délky lokální proměnné:
#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
extrn unit printf(string format, ...)

unit print_int_slice_backwards([slice_len]int slice) {
  for slice_len > 0 {
    slice_len = slice_len - 1;
    printf("slice[%d] = %d\n", slice_len, slice[slice_len]);
  }
}

unit main() {
  let n: int = 10;
  let arr: []int = make(int, n);
  print_int_slice_backwards(arr);
  release(arr);
}
```
]

== Neměnná globální hodnota
Použití funkce jako globální konstanty
#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
int max_buffer_size() { 1024 }

unit main() {
  let buffer: []int = make(int, max_buffer_size());
  release(buffer);
}
```
]

== Rekurze
Využití rekurze pro výpočet faktoriálu:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
int factorial(int n) {
    if n <= 1 {
        1
    } else {
        n * factorial(n - 1)
    }
}
```
]

== Cyklus
Využití návratové hodnody cyklu *for* pro iterativní výpočet faktoriálu:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
let a: int = {
               let n: int = 5;
               let res: int = n;
               for n > 1 {
                 n = n-1;
                 res = res * n
               }
             };
```
]


== Odkazy
Adresa proměnné se získá operátorem *^*, dereference operátorem *@*:
#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
int x: int = 42;
let p: ^int = ^x;
@p = 100;  # x = 100
```
]

Přečtení čísla z standardního vstupu:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
extrn int scanf(string format, ...)

int x: int = 0;
scanf("%d", ^x);
```
]


#pagebreak()
= Překladač
== Architektura
Překladač překládá zdrojový kód do GNU assembly pro architekturu x86_64.
Skládá se z těchto částí:
- *Lexikální analýza* - rozčlenění zdrojového kódu na tokeny
- *Syntaktická analýza* - sestavení abstraktního syntaktickýho stromu (AST) z tokenů
- *Vyhodnocení jmen* - projde strom a propojí identifikátory
- *Vyhodnocení typů* - projde strom a propaguje nahoru typy výrazů
- *Ověření typů* - ověření, jestli typy ve výrazech odpovídají očekávaným
- *Generátor kódu* - projde strom a vygeneruje odpovídající assembly

Výsledný assembly kód je přeložen pomocí GCC do spustitelného souboru.

== Volací konvence
Překladač generuje kód dodržující konvenci System V AMD64 ABI, která se používá na Linuxových systémech. Celočíselné argumenty jsou předávány nejprve šesti registry *%rdi*, *%rsi*, *%rdx*, *%rcx*, *%r8* a *%r9*, argumenty typu *float* nejprve osmi registry *%xmm0*, *%xmm1*, *%xmm2*, *%xmm3*, *%xmm4*, *%xmm5*, *%xmm6* a *%xmm7*. Další argumenty jsou předávány na zásobníku. Před voláním funkcí je zásobník zarovnán na 16 bajtů.

== Použití
Překladač používá konzolové rozhraní, které poskytuje následující argumenty:

- *-h* - vypíše argumenty překladače
- *-i* - umístění souboru se zdrojovým kódem k překladu
- *-o* - umístění přeloženého spustitelného souboru
- *-r* - přeložení programu a následné spuštění
- *-s* - umístění přeloženého assembly kódu
- *-a* - umístění AST grafu programu v graphviz .dot formátu
- *-t* - umístění vypsaných tokenů programu

Pro spuštění programů dostupných v *./examples* nebo zobrazení jejich ast lze použít program #link("https://github.com/casey/just")[#underline(stroke: (thickness: 0.1em, paint: purple))[just]].

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```sh
just --list
```
]

Spuštění programu kreslícího mandelbrotovu množinu a zobrazení jeho AST

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```sh
just r mandelbrot.ilang
just ad mandelbrot.ilang
```
]

Nebo manuálně:
#box(fill: rgb("#D3D3D3"), inset: 1em)[
```sh
go run cmd/compiler/main.go -i ./examples/mandelbrot.ilang -a mandelbrot.dot
dot -Tpng mandelbrot.dot -o mandelbrot.png
```
]

#pagebreak()
== Rozbor konkrétního programu
Následuje rozbor programu který implementuje buněčný automat #link("https://en.wikipedia.org/wiki/Rule_110")[#underline(stroke: (thickness: 0.1em, paint: purple))[rule110]].

Celý zdrojový kód je dostupný v *examples/rule110.ilang*

Deklarace dvou externích funkcí:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
extrn unit printf(string format, ...)
extrn unit scanf(string format, ...)
```
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
assembly:
```yasm
.extern printf
.extern scanf
```
]

Deklarace funkce která vypíše stav buněk:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
unit print_board([slice_len]int board) {
	let idx: int = 0;
	for idx < slice_len {
		printf(if !(board[idx] == 1) { " " } else { "#" });
		idx = idx + 1;
	};
	printf("\n");
}
```
]

Prolog funkce:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
assembly:
```yasm
print_board:
push %rbp
mov %rsp, %rbp
sub $32, %rsp # alokace lokálních proměnných
```
]

Přesun argumentů funkce do lokálních proměnných:


#box(fill: rgb("#D3D3D3"), inset: 1em)[
assembly:
```yasm
mov %rdi, -24(%rbp) # odkaz na pole
mov %rsi, -16(%rbp) # délka pole
mov %rsi, -8(%rbp) # délka pole dostupná v lokální proměnné
```
]

#pagebreak()

Přiřazení hodnoty *0* lokální proměnné *idx*:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
let idx: int = 0;
```
]


#box(fill: rgb("#D3D3D3"), inset: 1em)[
assembly:
```yasm
mov $0, %rax
mov %rax, -32(%rbp)
```
]

Cyklus *for*:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
for idx < slice_len {
  ...
};
```
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
assembly:
```yasm
mov $0, %rax # výchozí návratová hodnota cyklu
.label_1:
push %rax # uložení navratové hodnoty
# binární výraz
mov -32(%rbp), %rax # levá strana - lokální proměnná
push %rax
mov -8(%rbp), %rax # pravá strana - lokální proměnná
mov %rax, %rbx
pop %rax
cmp %rbx, %rax # binární operátor - menší než
setl %al
movzbq %al, %rax
cmp $1, %rax
jne .label_2 # pokud není pravda, skok na konec cyklu
pop %rax # smazání předchozí návratové hodnoty

  ... # tělo cyklu
  
jmp .label_1
.label_2:
pop %rax # načtení návratové hodnoty
```
]

#pagebreak()

Volání externí funkce *printf* s kondicí jako argumentem:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
printf(if !(board[idx] == 1) { " " } else { "#" });
```
]

#box(fill: rgb("#D3D3D3"), inset: 1em)[
assembly:
```yasm
mov -32(%rbp), %rax # hodnota indexu
push %rax
mov -24(%rbp), %rcx # odkaz na pole
pop %rdx
mov (%rcx, %rdx, 8), %rax # indexování odkazu na pole hodnotou
push %rax
mov $1, %rax # hodnota 1
mov %rax, %rbx
pop %rax
cmp %rbx, %rax # binární operátor - rovnost
sete %al
movzbq %al, %rax
cmp $0, %rax # kondice
sete %al
movzbq %al, %rax
cmp $1, %rax
je .label_3
jmp .label_4
.label_3: # větev if
lea .const_0(%rip), %rax # odkaz na řetězec
jmp .label_5
.label_4:  # větev else
lea .const_1(%rip), %rax # odkaz na řetězec
.label_5:
push %rax
mov 0(%rsp), %rax
mov %rax, %rdi # 1. argument funkce printf
add $8, %rsp # zarovnání zásobníku
mov $0, %rax # počet variadických argumentů
call printf@PLT # volání externí funkce
```
]

#pagebreak()

Funkce vracející hodnotu buňky v závislosti na jejím okolí:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
int rule110(int a, int b, int c) {
    let table: [8]int = [0, 1, 1, 1, 0, 1, 1, 0];
    let idx: int = (a << 2) || (b << 1) || c;
    table[idx]
}
```
]

Funkce která vypočítá nový stav buňek:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
unit next_iter([slice_len]int board, [next_len]int next_board) {
	let a: int = 0;
	let b: int = 0;
	let c: int = 0;
	let idx: int = 0;
	for idx < slice_len {
		if idx == 0 {
			a = board[(slice_len - 1)];
		} else {
			a = board[(idx - 1)];
		};
		b = board[idx];
		if idx == (slice_len-1) {
			c = board[0];
		} else {
			c = board[(idx + 1)];
		};
		let val: int = rule110(a,b,c);
		next_board[idx] = val;
		idx = idx+1;
	};
}
```
]

Funkce která vypočítá a vypíše po *n* sobě jdoucích stavů buněk:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
unit print_n_iterations([board_len]int board, [next_len]int next_board, int iters) {
	for iters > 0 {
		let tmp: []int = board;
		next_iter(board, next_board);
		print_board(next_board);
		tmp = board;
		board = next_board;
		next_board = tmp;
		iters = iters - 1;
	}
}
```
]

#pagebreak()

Funkce k přečtení čísla od uživatele ze standardního vstupu:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
int read_number_from_stdin(string prompt) {
	let number: int = 0;
	printf(prompt);
	scanf("%d", ^number);
	number
}
```
]

Vstupní bod programu:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```ilang
int main() {
	let size: int = read_number_from_stdin("board size: ");
	let board: []int = make(int, size);
	let next_board: []int = make(int, size);
	board[size-1] = 1;
	print_board(board);
	print_n_iterations(board, next_board, size-1);
	release(board);
	release(next_board);
	0
}
```
]

Příklad výstupu pro *n* = 10:

#box(fill: rgb("#D3D3D3"), inset: 1em)[
```
board size: 10
         #
        ##
       ###
      ## #
     #####
    ##   #
   ###  ##
  ## # ###
 ####### #
##     ###
```
]
