# Golang-Challenge
Challenge test

We ask that you complete the following challenge to evaluate your development skills.

## The Challenge
Finish the implementation of the provided Transparent Cache package.

## Show your work

1.  Create a **Private** repository and share it with the recruiter ( please dont make a pull request, clone the private repository and create a new private one on your profile)
2.  Commit each step of your process so we can follow your thought process.
3.  Give your interviewer access to the private repo

## What to build
Take a look at the current TransparentCache implementation.

You'll see some "TODO" items in the project for features that are still missing.

The solution can be implemented either in Golang or Java ( but you must be able to read code in Golang to realize the exercise ) 

Also, you'll see that some of the provided tests are failing because of that.

The following is expected for solving the challenge:
* Design and implement the missing features in the cache
* Make the failing tests pass, trying to make none (or minimal) changes to them
* Add more tests if needed to show that your implementation really works
 
## Deliverables we expect:
* Your code in a private Github repo
* README file with the decisions taken and important notes

## Time Spent
We suggest not to spend more than 2 hours total, which can be done over the course of 2 days.  Please make commits as often as possible so we can see the time you spent and please do not make one commit.  We will evaluate the code and time spent.
 
What we want to see is how well you handle yourself given the time you spend on the problem, how you think, and how you prioritize when time is insufficient to solve everything.

Please email your solution as soon as you have completed the challenge or the time is up.

## Decisions taken

* To parallelize calls to GetPriceFor, I decided to use the Go feature errgroup.WithContext which provides synchronization, error propagation, and Context cancelation for groups of goroutines. The derived Context is canceled the first time a function passed to Go returns a non-nil error or the first time Wait returns, whichever occurs first.

* Changed prices vaue type from float64 to Data struct type, which contains the same float64 value itfself and time expiration for each item that is added to prices. Time expiration will be defined by a number that is a result of adding maxAge to time.Now, so I can compare later (when I read same key) to see if the new requested time will be lower that expiration value.

* Added a sync.RWMutex to TransparentCache (reader/writer mutual exclusion lock) to avoid race conditions when parallelized GetPriceFor calls, which access to prices map for reading and writing at the same time from different goroutinies.
For this reason, I've created Read and Write methods to wrap get and set cache functionality with locks for concurrent read/writes. 

* Checking expiration key is inside Read method, not in GetPricesFor as the exercise was at the first time, because if not, every this check must be place in every read key request all over the code. Now, Read method only return the value only if key exists and isn't expired. Also, if the key is expired, I decided to delete it there to avoid storing unused keys, until it will be written again.

* I took advantage you gave me to make a small change to the test you provided me, and I added mutual exclusion on: 
```m.numCalls++``` to avoid race conditions when testing parallel request on GetPricesFor with errgroup.

* I added 3 more unit test for checking the correctness of my code.

# Improvement for the future

As I mentioned in the code, deleting the expired keys in the Read method when I check that they has expired, is not a good solution, since it only eliminates the keys that are consulted again, but those that are not, are stored in the map consuming memory and I know that thos values aren't valid any more.
The exercise didn't ask to do it, but I'd clearly do another erase mechanism when the TTL of each key is met. The simplest is to define a task that runs every certain time configured and delete those keys that are expired. 