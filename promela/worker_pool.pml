#define JOBS 5
#define WORKERS 3

chan jobs = [0] of { byte };
chan results = [JOBS] of { byte };
bool jobsClosed = false;
bool resultsClosed = false;
byte workersDone = 0;
bool completed[JOBS];
byte completedCount = 0;
byte writers = 0;

active proctype Producer()
{
	byte job;
	for (job : 0 .. JOBS - 1) {
		jobs ! job
	}
	jobsClosed = true
}

active [WORKERS] proctype Worker()
{
	byte job;
	do
	:: jobs ? job ->
		assert(job < JOBS);
		results ! job
	:: jobsClosed -> break
	od;
	/* Abstraccion de la actualizacion sincronizada de WaitGroup.Done. */
	atomic { workersDone++ }
}

active proctype Closer()
{
	(workersDone == WORKERS);
	resultsClosed = true
}

/* Un unico consumidor escribe el arreglo, igual que en Go. */
init
{
	byte job;
	byte index;
	do
	:: results ? job ->
		writers++;
		assert(writers == 1);
		assert(job < JOBS);
		assert(!completed[job]);
		completed[job] = true;
		completedCount++;
		assert(completedCount <= JOBS);
		writers--
	:: (resultsClosed && len(results) == 0) -> break
	od;
	assert(completedCount == JOBS);
	for (index : 0 .. JOBS - 1) {
		assert(completed[index])
	}
}

ltl all_jobs_complete { <> (completedCount == JOBS && resultsClosed) }
