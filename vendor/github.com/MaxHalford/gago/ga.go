package gago

import (
	"errors"
	"log"
	"math/rand"
	"sort"
	"time"
)

// A GA contains population which themselves contain individuals.
type GA struct {
	// Required fields
	NewGenome NewGenome `json:"-"`
	NPops     int       `json:"-"` // Number of Populations
	PopSize   int       `json:"-"` // Number of Individuls per Population
	Model     Model     `json:"-"`

	// Optional fields
	ParallelEval bool         `json:"-"` // Evaluate Individuals in parallel or not
	NBest        int          `json:"-"` // Length of HallOfFame
	Migrator     Migrator     `json:"-"`
	MigFrequency int          `json:"-"` // Frequency at which migrations occur
	Speciator    Speciator    `json:"-"`
	Logger       *log.Logger  `json:"-"`
	Callback     func(ga *GA) `json:"-"`
	RNG          *rand.Rand   `json:"-"`

	// Fields generated at runtime
	Populations Populations   `json:"populations"`
	HallOfFame  Individuals   `json:"hall_of_fame"` // Sorted best Individuals ever encountered
	Age         time.Duration `json:"duration"`     // Duration during which the GA has been evolved
	Generations int           `json:"generations"`  // Number of generations the GA has been evolved
}

// Validate the parameters of a GA to ensure it will run correctly; some
// settings or combination of settings may be incoherent during runtime.
func (ga GA) Validate() error {
	// Check the NewGenome presence
	if ga.NewGenome == nil {
		return errors.New("NewGenome cannot be nil")
	}
	// Check the number of populations is higher than 0
	if ga.NPops < 1 {
		return errors.New("NPops should be higher than 0")
	}
	// Check the number of individuals per population is higher than 0
	if ga.PopSize < 1 {
		return errors.New("PopSize should be higher than 0")
	}
	// Check the model presence
	if ga.Model == nil {
		return errors.New("Model cannot be nil")
	}
	// Check the model is valid
	var modelErr = ga.Model.Validate()
	if modelErr != nil {
		return modelErr
	}
	// Check the migration frequency if a Migrator has been provided
	if ga.Migrator != nil && ga.MigFrequency < 1 {
		return errors.New("MigFrequency should be strictly higher than 0")
	}
	// Check the speciator is valid if it has been provided
	if ga.Speciator != nil {
		if specErr := ga.Speciator.Validate(); specErr != nil {
			return specErr
		}
	}
	// No error
	return nil
}

// Find the best current Individual in each population and then compare the best
// overall Individual to the current best Individual. The Individuals in each
// population are expected to be sorted.
func updateHallOfFame(hof Individuals, indis Individuals, rng *rand.Rand) {
	var k = len(hof)
	// Start by finding the current best Individual
	for _, indi := range indis[:min(k, len(indis))] {
		// Find if and where the Individual should fit in the hall of fame
		var (
			f = func(i int) bool { return indi.Fitness < hof[i].Fitness }
			i = sort.Search(k, f)
		)
		if i < k {
			// Shift the hall of fame to the right
			copy(hof[i+1:], hof[i:])
			// Insert the new Individual
			hof[i] = indi.Clone(rng)
		}
	}
}

// Initialized indicates if the GA has been initialized or not.
func (ga GA) Initialized() bool {
	return len(ga.Populations) == ga.NPops && len(ga.HallOfFame) == ga.NBest
}

// Initialize each population in the GA and assign an initial fitness to each
// individual in each population. Running Initialize after running Evolve will
// reset the GA entirely.
func (ga *GA) Initialize() error {
	// Check the NBest field
	if ga.NBest < 1 {
		ga.NBest = 1
	}
	// Initialize the random number generator if it hasn't been set
	if ga.RNG == nil {
		ga.RNG = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	ga.Populations = make([]Population, ga.NPops)
	var err error
	for i := range ga.Populations {
		// Generate a Population
		ga.Populations[i] = newPopulation(ga.PopSize, ga.NewGenome, ga.RNG)
		// Evaluate it's Individuals
		err = ga.Populations[i].Individuals.Evaluate(ga.ParallelEval)
		if err != nil {
			return err
		}
		// Sort it's Individuals
		ga.Populations[i].Individuals.SortByFitness()
		// Log current statistics if a logger has been provided
		if ga.Logger != nil {
			ga.Populations[i].Log(ga.Logger)
		}
	}
	// Initialize HallOfFame
	ga.HallOfFame = make(Individuals, ga.NBest)
	for i := range ga.HallOfFame {
		ga.HallOfFame[i] = ga.Populations[0].Individuals[0].Clone(ga.RNG)
	}
	for _, pop := range ga.Populations {
		updateHallOfFame(ga.HallOfFame, pop.Individuals, pop.RNG)
	}
	// Execute the callback if it has been set
	if ga.Callback != nil {
		ga.Callback(ga)
	}
	return nil
}

// Evolve each population in the GA. The population level operations are done
// in parallel with a wait group. After all the population operations have been
// run, the GA level operations are run.
func (ga *GA) Evolve() error {
	var start = time.Now()

	// Check the GA has been initialized
	if !ga.Initialized() {
		return errors.New("The GA has not been initialized")
	}

	// Migrate the individuals between the populations if there are at least 2
	// Populations and that there is a migrator and that the migration frequency
	// divides the generation count
	if len(ga.Populations) > 1 && ga.Migrator != nil && ga.Generations%ga.MigFrequency == 0 {
		ga.Migrator.Apply(ga.Populations, ga.RNG)
	}

	var f = func(pop *Population) error {
		var err error
		// Apply speciation if a positive number of species has been specified
		if ga.Speciator != nil {
			err = pop.speciateEvolveMerge(ga.Speciator, ga.Model)
			if err != nil {
				return err
			}
		} else {
			// Else apply the evolution model to the entire population
			err = ga.Model.Apply(pop)
			if err != nil {
				return err
			}
		}
		// Evaluate and sort
		err = pop.Individuals.Evaluate(ga.ParallelEval)
		if err != nil {
			return err
		}
		pop.Individuals.SortByFitness()
		// Record time spent evolving
		pop.Age += time.Since(start)
		pop.Generations++
		// Log current statistics if a logger has been provided
		if ga.Logger != nil {
			pop.Log(ga.Logger)
		}
		return err
	}

	var err = ga.Populations.Apply(f)
	if err != nil {
		return err
	}
	// Update HallOfFame
	for _, pop := range ga.Populations {
		updateHallOfFame(ga.HallOfFame, pop.Individuals, pop.RNG)
	}
	// Execute the callback if it has been set
	if ga.Callback != nil {
		ga.Callback(ga)
	}
	ga.Age += time.Since(start)
	ga.Generations++
	// No error
	return nil
}

func (pop *Population) speciateEvolveMerge(spec Speciator, model Model) error {
	var (
		species, err = spec.Apply(pop.Individuals, pop.RNG)
		pops         = make([]Population, len(species))
	)
	if err != nil {
		return err
	}
	// Create a subpopulation from each specie so that the evolution Model can
	// be applied to it.
	for i, specie := range species {
		pops[i] = Population{
			Individuals: specie,
			Age:         pop.Age,
			Generations: pop.Generations,
			ID:          randString(len(pop.ID), pop.RNG),
			RNG:         pop.RNG,
		}
		err = model.Apply(&pops[i])
		if err != nil {
			return err
		}
	}
	// Merge each species back into the original population
	var i int
	for _, subpop := range pops {
		copy(pop.Individuals[i:i+len(subpop.Individuals)], subpop.Individuals)
		i += len(subpop.Individuals)
	}
	return nil
}
