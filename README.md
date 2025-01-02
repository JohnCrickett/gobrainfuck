# Go Brainfuck
A Brainfuck interpreter written in Go for [Coding Challenges](https://codingchallenges.fyi/challenges/intro/).

## About Brainfuck

Brainfuck is an esoteric programming language that was created by Urban Müller in 1993. 
Apparently the intention was to be able to create a language that allowed for the creation of a tiny compiler. 
The first compiler being just 296 bytes, which later shrank to 240 bytes. 

It is an example of a [Turing-complete](https://samwho.dev/turing-machines/#what-does-it-mean-to-be-turing-complete) 
language, therefore like any Turing-complete language, Brainfuck is theoretically capable of computing any 
computable function or simulating any other computational model if given access to an unlimited amount of memory 
and time. In reality, it's not practical for real-world usage as it provides almost no abstraction making it tedious 
to develop anything other than toy programs with it.

### The Brainfuck Language

Brainfuck operates on an array of memory cells, each initialised to zero. 
In the original implementation, the array was 30,000 cells long. 
There is an instruction pointer than points to the command to be executed, 
which then moves forward after the instruction has been executed.
There is a data pointer that initially points to the first memory cell. 

The language consists of eight commands:

| Command | Description                                                       |
|---------|-------------------------------------------------------------------|
| `>`     | Move the pointer to the right                                     |
| `<`     | Move the pointer to the left                                      |
| `+`     | Increment the memory cell at the pointer                          |
| `-`     | Decrement the memory cell at the pointer                          |
| `.`     | Output the character signified by the cell at the pointer         |
| `,`     | Input a character and store it in the cell at the pointer         |
| `[`     | Jump past the matching ] if the cell at the pointer is 0          |
| `]`     | Jump back to the matching [ if the cell at the pointer is nonzero |

All characters other than `><+-.,[]` are considered to comments and are ignored.

## Building and Running Go Brainfuck

### Building Go Brainfuck

```bash
go build .
```

### Running Brainfuck Code From File

The interpeter will load and run the first provided filename. There are some example programs in the bf-scripts 
directory that can be used for testing / experimentation.


```bash
gobrainfuck bf-scripts/hello.bf
```

### Running Brainfuck Code In the REPL

The REPL expects the line to be a full, valid program.

```bash
 % gobrainfuck                           
BF> +++++++++++[>++++++>+++++++++>++++++++>++++>+++>+<<<<<<-]>++++++.>++.+++++++..+++.>>.>-.<<-.<.+++.------.--------.>>>+.>-.
Hello, World!
BF> quit
```

### Command Line Options

The interpterer allows you to specify the number of cells (size of the memory) on the command line:

```bash
 % ./gobrainfuck -help
Usage of ./gobrainfuck:
  -cells int
        Number of cells to use (default 30000)
```