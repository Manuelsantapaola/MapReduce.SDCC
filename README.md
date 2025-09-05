# MapReduce.SDCC
Realize a distributed application that solves the sorting problem using MapReduce
Master-Worker Architecture:

The application is structured following a master-worker architecture, where a Master coordinates the work and assigns tasks to Workers.
The master assigns both the map and reduce tasks to the workers.

MapReduce Phases:

Map: Each worker receives a portion of the data (called a chunk) from the master to sort. The worker performs sorting on the assigned data chunk.

Shuffle: The workers that have completed the map task send their intermediate results to other workers responsible for the reduce phase. This involves partitioning the intermediate data.

Reduce: The workers receiving the intermediate results (the sorted data from map workers) perform a merge of these results to produce the final sorted output. Each reduce worker writes the final result to a file.

Synchronization Points:

The reduce phase cannot start until all map workers have completed their processing. This requires a synchronization point between the map and reduce phases.

Simplifying Assumptions:

No failure occurs in the master or any worker during execution.

The set of workers is defined at the beginning and does not change during execution.

Communication ports are predefined in a configuration file.
