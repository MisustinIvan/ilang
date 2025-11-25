#set text(font: "Iosevka NF")

= Téma maturitní práce
== Stručný popis
Jednoduchý překládaný programovací jazyk implementován v jazyce Go. Cílem je si vyzkoušet návrh a implementaci vlastního jazyka od gramatiky po překlad. Jazyk bude úmyslně minimalistický, ale plně funkční, aby bylo možné demonstrovat klíčové principy tvorby jazka.

== Zaměření jazyka
Jazyk bude ze syntaktické stránky z velké části inspirován jazyky C a Rust. Bude orientován na mix mezi imperativním a deklarativním stylem programování.
Samozřejmá je podpora základních funkcí jazyka:
- práce s proměnými
- práce s výrazy
- definice funkcí a jejich volání
- základní řídící struktury
- různé datové typy

Mezi plánované(nejisté) části patří:
- tvorba vlastních datových typů
- struktury
- metody na datových typech
- více souborů zdrojového kódu najednou, případně moduly

Cílem je jednoduchý jazyk, který bude možno rozšířit o složitější funkce pokud bude čas.

== Plánované části překladače
- *Lexer* - provádí lexickou analýzu a dělí zdrojový kód na tokeny
- *Parser* - převede tokeny na AST
- *Name Resolver* - vyhledání symbolů
- *Type Resolver* - určení typů jednotlivých částí AST
- *Type Checker* - zkontroluje jestli se typy shodují s očekávanými
- (nejisté)*Optimizer* - optimalizuje AST(constant folding, dead code elimination, constant propagation)
- (nejisté)*IR Generator* - generuje IR pro další optimalizaci
- *Code generation* - generuje assembly z AST nebo IR

== Odevzdaný obsah
- všechen zdrojový kód
- přeloženou verzi překladače pro distribuci
- specifikaci jazyka
- formální gramatiku jazyka v EBNF
- dokumentaci k překladači
