# Optimising The Interpreter

Now that the Interpreter works it's a fun side project to optimise it. So before we start let's get a baseline.

```bash
% time ./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1
./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1  11.48s user 0.03s system 99% cpu 11.552 total
% time ./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1
./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1  11.50s user 0.03s system 99% cpu 11.627 total
% time ./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1
./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1  11.48s user 0.03s system 99% cpu 11.591 total
```

Fairly consistent around 11.48s. For the first optimisation we'll compress repeated instructions. For example
`>>>>>` will become `>` with an operand of 5. 

```bash
% time ./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1
./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1  5.81s user 0.02s system 99% cpu 5.843 total
% time ./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1
./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1  5.81s user 0.02s system 99% cpu 5.849 total
% time ./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1
./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1  5.85s user 0.03s system 98% cpu 5.962 total
```

Fairly consistent around 5.81s.

For the next level of optimisation I'm going to identify common 'functions', i.e. repeated patterns and create single
OpCodes for them in the bytecode. We can identify the patterns during compilation and we can work out how often they're
used by logging them from an execution.

This is enabled by setting Profile to `true`. After a couple of optimisations it becomes:

```bash
% time ./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1
./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1  4.59s user 0.03s system 84% cpu 5.497 total
% time ./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1
./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1  4.50s user 0.02s system 99% cpu 4.532 total
% time ./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1
./gobrainfuck bf-scripts/mandelbrot.bf > /dev/null 2>&1  4.48s user 0.02s system 99% cpu 4.530 total
```


