#define JOBS 5
#define WORKERS 3
#define STOP JOBS

chan jobs = [JOBS + WORKERS] of { byte };
chan results = [JOBS] of { byte };
bool completed[JOBS];
byte completedCount = 0;
bool inCritical = false;

active [WORKERS] proctype Worker()
{
	byte job;
	do
	:: jobs ? job ->
		if
		:: job == STOP -> break
		:: else ->
			atomic {
				assert(job < JOBS);
				assert(!completed[job]);
				assert(!inCritical);
				inCritical = true;
				completed[job] = true;
				completedCount++;
				inCritical = false;
			}
			results ! job
		fi
	od
}

init
{
	byte job;
	atomic {
		for (job : 0 .. JOBS - 1) {
			jobs ! job
		}
		for (job : 0 .. WORKERS - 1) {
			jobs ! STOP
		}
	}
	d_step {
		assert(completedCount <= JOBS);
	}
}

ltl all_jobs_complete { [] (completedCount < JOBS -> <> (completedCount == JOBS)) }