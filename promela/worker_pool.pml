#define JOBS 5
#define WORKERS 3

chan jobs = [JOBS] of { byte };
chan results = [JOBS] of { byte };
bool completed[JOBS];
byte completedCount = 0;

active [WORKERS] proctype Worker()
{
	byte job;
	do
	:: jobs ? job ->
		atomic {
			assert(job < JOBS);
			assert(!completed[job]);
			completed[job] = true;
			completedCount++;
		}
		results ! job
	od
}

init
{
	byte job;
	atomic {
		for (job : 0 .. JOBS - 1) {
			jobs ! job
		}
	}
	d_step {
		assert(completedCount <= JOBS);
	}
}

ltl all_jobs_complete { [] (completedCount == JOBS -> <> (completedCount == JOBS)) }