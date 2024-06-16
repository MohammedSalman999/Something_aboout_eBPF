**How to Use the Go Program with eBPF to Drop TCP Packets on a Port**

This guide will walk you through compiling the eBPF C code into an object file and using a Go program to load and run the eBPF program. The eBPF program will drop TCP packets destined for a specific port.

**Prerequisites:**

- **Linux Environment:** This setup requires a Linux environment where eBPF programs can be loaded and attached.
- **clang:** Ensure you have clang installed to compile the eBPF C code into an object file.
- **Go:** Install Go (Golang) on your system. You can download it from [golang.org](https://golang.org).

**Step-by-Step Instructions:**

1. **Compile the eBPF C Code into an Object File**

   - **Open a Terminal:** Open your terminal application.

   - **Navigate to the Directory:** Change directory (`cd`) to where your `eBPF.c` file is located.

   - **Compile the C Code:** Use clang to compile the `eBPF.c` into an eBPF object file (`eBPF.o`):

     ```sh
     clang -O2 -target bpf -c eBPF.c -o eBPF.o
     ```

     This command compiles `eBPF.c` into `eBPF.o` using the `bpf` target.

2. **Run the Go Program**

   - **Download the Go Program:** Ensure you have downloaded the Go program (`main.go`) provided in this repository.

   - **Open the Go Program:** Open `main.go` in a text editor to review or modify if necessary.

   - **Run the Go Program:** In the terminal, navigate to the directory containing `main.go` and `eBPF.o`.

   - **Build the Go Program:** Build the Go program using `go build`:

     ```sh
     go build -o tcp_drop main.go
     ```

   - **Execute the Go Program:** Run the compiled Go program with superuser privileges (`sudo`), providing the desired TCP port number to drop packets (default is `4040`):

     ```sh
     sudo ./tcp_drop -port 4040
     ```

     Replace `4040` with the desired TCP port number. This command attaches the compiled eBPF program to the network interface (`eth0` by default) to drop TCP packets on the specified port.

**Additional Notes:**

- **Customization:** You can modify `eBPF.c` to change the port number or add additional filtering logic in the eBPF program.
  
- **Cleanup:** Use `Ctrl + C` to terminate the Go program and detach the eBPF program cleanly.

**Troubleshooting:**

- **Permissions:** If you encounter permission errors, ensure you have superuser (`sudo`) privileges to attach the eBPF program.
  
- **Dependencies:** Ensure `clang` and Go are properly installed and accessible from your terminal.


