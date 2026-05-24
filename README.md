# RISC-V Vector Tests Generator

## About

This repository hosts unit tests generator for the RISC-V vector extension.

## Features

- Similar to [riscv-tests](https://github.com/riscv-software-src/riscv-tests), simple and easy to use
- Self-verification by default, Co-simulator friendly
- User-mode and machine-mode binaries
- TestFloat3 integration
- Support RV32 and RV64
- Test SEW from e8 to e64
- Test LMUL from mf8 to m8
- Support VLEN from 32 to 65536 (VLEN=32 requires XLEN=32)
- Support varies sub-extensions: Zvfh, Zvbb, Zvbc, Zvkg, Zvkned, Zvknha, Zvksed, Zvksh, Zvfbfmin and Zvfbfwma
- Configurable, see `make help`

## Limitations

- All the tests are per instruction, and it is done more or less in the same fashion
- Lack of tail/mask agnostic support (i.e. ta/ma)
- Lack of fault-only-first testing
- Lack of vstart testing
- Lack of register group overlap testing
- ... and more

Overall, there are no coverage statistics or guarantees.

## How it works

The Spike simulator is known as the RISC-V gold standard simulator, and although we don't know how Spike is tested, it does fully support the V extension. So we added [a custom special instruction](https://github.com/ksco/riscv-vector-tests/blob/6a23892a5ab0cc72f4867cc95186b3528c99c2a0/pspike/pspike.cc#L20) to Spike, and for any test, let it automatically generate a reference result for that test. This way, we generate tests for all instructions almost automatically. Under this framework, all we have to do is write a [simple config file for each instruction](configs/).

## Prerequisite

1. `riscv64-unknown-elf-gcc` with RVV 1.0 support
2. The Spike simulator
3. Golang 1.19+
4. `riscv-pk` if you need to generate user-mode binaries

## Cloning

This repository uses the `riscv-test-env` repository as a submodule.

```
git clone --recurse-submodules https://github.com/chipsalliance/riscv-vector-tests
```

## How to use

```
make all -j$(nproc)
```

> If you have problems compiling, please refer to the build steps in [build-and-test.yml](.github/workflows/build-and-test.yml).

After `make all`, you will find all the generated tests in `out/v[vlen]x[xlen][mode]/bin/stage2/`.

For more advanced options, run `make help`.

> Note: [single/single.go](single/single.go) generates tests directly from stage 1, suitable for targets with co-simulators (or simply use `TEST_MODE=cosim` if you're lazy).

### Nix package

This repository also provides a nix derivation with the following output provided:

- `${riscv-vector-test}/bin/*`: Generator binaries
- `${riscv-vector-test}/include/*`: Necessary headers for runtime usage
- `${riscv-vector-test}/configs/*`: Necessary runtime configs

## License

This project uses third-party projects, and the licenses of these projects are attached to the corresponding directories.

The code for this project is distributed under the Apache License Version 2.0.

The “RISC-V” trade name is a registered trademark of RISC-V International.
