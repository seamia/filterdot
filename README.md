# filterdot
filter out parts of your (directed) .dot graph that you are not interested in

## show me
go from this:
<p align="center">
  <img src="https://raw.githubusercontent.com/seamia/filterdot/main/.media/before.svg">
</p>
to this:
<p align="center">
  <img src="https://raw.githubusercontent.com/seamia/filterdot/main/.media/after.svg">
</p>

## why
sometimes/often the whole graph is too big to be usable, especially if you are interested in a small part of it.

this simple utility allows you to produce a smaller graph based on a list of nodes (and their descendents) to be included ("+") 
and a list of nodes to be excluded ("-")

## how
CLI usage:
```
./filterdot Original.dot Resulting.dot +123 +783 -245
```

## build
1. install `go`:
    ```
    https://golang.org/doc/install
   ```
2. create a directory and "go there"
3. get the source:
    ```
    git clone https://github.com/seamia/filterdot.git
   ```
4. build
    ```
   go build
    ```
