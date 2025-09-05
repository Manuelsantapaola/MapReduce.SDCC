# MapReduce.SDCC
# 🗂️ Distributed Sorting with MapReduce

## 📌 Project Overview

This project implements a **distributed sorting algorithm** using the **MapReduce** programming model within a **Master-Worker architecture**.

The application is designed to divide a large sorting task across multiple workers, enabling parallel processing and scalability.

---

## 🧠 Architecture: Master-Worker

- The system follows a **Master-Worker architecture**.
- The **Master** node is responsible for:
  - Coordinating the execution
  - Assigning **Map** and **Reduce** tasks to workers
- **Worker** nodes perform the actual sorting and merging tasks.

---

## 🗃️ MapReduce Phases

### 🔹 1. Map Phase
- Each worker receives a **chunk** of unsorted data from the master.
- The worker **sorts its chunk locally**.

### 🔹 2. Shuffle Phase
- After sorting, each map worker sends its sorted data to one or more reduce workers.
- Data is **partitioned** appropriately to ensure correct final ordering.

### 🔹 3. Reduce Phase
- Reduce workers **merge** the sorted chunks received from the map workers.
- Each reduce worker produces a **final sorted output file**.

---

## ⏳ Synchronization

- The **Reduce phase starts only after all Map workers have completed**.
- A **synchronization barrier** ensures that Reduce tasks do not begin prematurely.

---

## ⚙️ Simplifying Assumptions

- **No failures** are expected in either the master or workers during execution.
- The set of workers is **fixed at startup** and remains unchanged.
- **Communication ports** for all workers are predefined in a configuration file.

---

## 🛠️ Technologies and Tools

- Programming Language: *(add your language, e.g., Python, Java, C++)*
- Network Communication: *(e.g., sockets, gRPC, REST)*
- Data Format: *(e.g., plain text, JSON, binary)*

---

## 📁 Repository Structure


