package optimus

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/MaxHalford/gago"
	"go.uber.org/zap"
)

// Orders is a slice of market orders.
//
// Introduced to simplify cloning operation.
type Orders []*MarketOrder

// Clone performs a shallow copy of this orders slice.
func (m Orders) Clone() Orders {
	cp := make([]*MarketOrder, len(m))
	copy(cp, m)
	return cp
}

// Shuffle randomly shuffles this orders slice.
// This is the modern version of the Fisher–Yates shuffle, because we must
// support go1.9.
func (m Orders) Shuffle(rng *rand.Rand) {
	for i := range m {
		j := rng.Intn(i + 1)
		m[i], m[j] = m[j], m[i]
	}
}

type ordersGenome struct {
	knapsack *Knapsack
	orders   Orders // Constant.
}

// IsAdapted determines whether this genome is capable of survival.
func (m *ordersGenome) IsAdapted() bool {
	return len(m.knapsack.Plans()) > 0
}

type packedOrdersGenome struct {
	ordersGenome
	candidates Orders
}

type NewGenomeLab func(knapsack *Knapsack, orders []*MarketOrder) gago.NewGenome

func NewPackedOrdersNewGenome(knapsack *Knapsack, orders []*MarketOrder) gago.NewGenome {
	return func(rng *rand.Rand) gago.Genome {
		populationSize := rng.Int() % len(orders)
		candidates := Orders(orders).Clone()[:populationSize]
		candidates.Shuffle(rng)

		genome := &packedOrdersGenome{
			ordersGenome: ordersGenome{
				knapsack: knapsack.Clone(),
				orders:   orders,
			},
			candidates: candidates,
		}

		return genome
	}
}

// Pack packs the specified knapsack with evolved individuals.
//
// The fitness criteria must be checked before calling, otherwise an error
// can be returned.
func (m *packedOrdersGenome) Pack(knapsack *Knapsack) error {
	for _, order := range m.candidates {
		if err := knapsack.Put(order.Order); err != nil {
			return err
		}
	}

	return nil
}

func (m *packedOrdersGenome) Evaluate() float64 {
	knapsack := m.knapsack.Clone()

	for _, order := range m.candidates {
		switch err := knapsack.Put(order.Order); err {
		case nil:
		case errExhausted:
			return 0.0
		default:
			return math.NaN()
		}
	}

	if len(knapsack.Plans()) == 0 {
		return 0.0
	}

	price := float64(knapsack.Price().GetPerSecond().Unwrap().Uint64()) / 1e18

	// We want to minimize the fitness, hence the reversing.
	return -price
}

func (m *packedOrdersGenome) Mutate(rng *rand.Rand) {
	if len(m.candidates) == 0 {
		m.candidates = append(m.candidates, m.orders[rng.Int()%len(m.orders)])
		return
	}

	if m.IsAdapted() {
		// Add random order.
		m.candidates = append(m.candidates, m.orders[rng.Int()%len(m.orders)])
	} else {
		// Delete random order.
		id := rng.Int() % len(m.candidates)
		m.candidates = append(m.candidates[:id], m.candidates[id+1:]...)
	}
}

func (m *packedOrdersGenome) Crossover(genome gago.Genome, rng *rand.Rand) {
	if len(m.candidates) == 0 && len(genome.(*packedOrdersGenome).candidates) == 0 {
		return
	}

	if len(m.candidates) == 0 {
		// Add random order from the other parent.
		id := rng.Int() % len(genome.(*packedOrdersGenome).candidates)
		m.candidates = append(m.candidates, genome.(*packedOrdersGenome).candidates[id])
		return
	}

	if len(genome.(*packedOrdersGenome).candidates) == 0 {
		id := rng.Int() % len(m.candidates)
		genome.(*packedOrdersGenome).candidates = append(genome.(*packedOrdersGenome).candidates, m.candidates[id])
		return
	}

	for i := 0; i < 10; i++ {
		j := rng.Int() % len(m.candidates)
		k := rng.Int() % len(genome.(*packedOrdersGenome).candidates)

		m.candidates[j], genome.(*packedOrdersGenome).candidates[k] = genome.(*packedOrdersGenome).candidates[k], m.candidates[j]
	}
}

func (m *packedOrdersGenome) Clone() gago.Genome {
	return &packedOrdersGenome{
		ordersGenome: ordersGenome{
			knapsack: m.knapsack.Clone(),
			orders:   m.orders,
		},
		candidates: m.candidates.Clone(),
	}
}

type DecisionVec []float64

func (m DecisionVec) Clone() DecisionVec {
	cp := make(DecisionVec, len(m))
	copy(cp, m)
	return cp
}

type decisionOrdersGenome struct {
	ordersGenome
	decisions DecisionVec
}

func NewDecisionOrdersNewGenome(knapsack *Knapsack, orders []*MarketOrder) gago.NewGenome {
	return func(rng *rand.Rand) gago.Genome {
		return &decisionOrdersGenome{
			ordersGenome: ordersGenome{
				knapsack: knapsack.Clone(),
				orders:   orders,
			},
			decisions: make(DecisionVec, len(orders)),
		}
	}
}

func (m *decisionOrdersGenome) Pack(knapsack *Knapsack) error {
	for id, probability := range m.decisions {
		if probability > 0.5 {
			if err := knapsack.Put(m.orders[id].Order); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *decisionOrdersGenome) Evaluate() float64 {
	knapsack := m.knapsack.Clone() // TODO: Is this required?

	for id, probability := range m.decisions {
		if probability > 0.5 {
			order := m.orders[id].Order

			switch err := knapsack.Put(order); err {
			case nil:
			case errExhausted:
				return 0.0
			default:
				return math.NaN()
			}
		}
	}

	if len(knapsack.Plans()) == 0 {
		return 0.0
	}

	price := float64(knapsack.Price().GetPerSecond().Unwrap().Uint64()) * 1e-18

	// We want to minimize the fitness, hence the reversing.
	return -price
}

func (m *decisionOrdersGenome) Mutate(rng *rand.Rand) {
	if rng.Float64() < 1.0/8.0 {
		id := rng.Int() % len(m.decisions)

		if m.decisions[id] > 0.5 {
			m.decisions[id] = 0.0
		} else {
			m.decisions[id] = 1.0
		}
	}

	if rng.Float64() < 1.0/16.0 {
		id := rng.Int() % len(m.decisions)

		if m.decisions[id] > 0.5 {
			m.decisions[id] = 0.0
		} else {
			m.decisions[id] = 1.0
		}
	}
}

func (m *decisionOrdersGenome) Crossover(genome gago.Genome, rng *rand.Rand) {
	gago.CrossGNXFloat64(m.decisions, genome.(*decisionOrdersGenome).decisions, len(m.decisions)/10, rng)
}

func (m *decisionOrdersGenome) Clone() gago.Genome {
	return &decisionOrdersGenome{
		ordersGenome: ordersGenome{
			knapsack: m.knapsack.Clone(),
			orders:   m.orders,
		},
		decisions: m.decisions.Clone(),
	}
}

type GenomeConfig struct {
	NewGenomeLab `json:"-"`
	Type         string
}

func (m *GenomeConfig) MarshalText() (text []byte, err error) {
	return []byte(m.Type), nil
}

func (m *GenomeConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ty string
	if err := unmarshal(&ty); err != nil {
		return err
	}

	switch ty {
	case "packed":
		m.NewGenomeLab = NewPackedOrdersNewGenome
	case "decision":
		m.NewGenomeLab = NewDecisionOrdersNewGenome
	default:
		return fmt.Errorf("unknown genome: %s", m.Type)
	}

	m.Type = ty

	return nil
}

type GeneticModelConfig struct {
	Genome         GenomeConfig  `yaml:"genome"`
	PopulationSize int           `yaml:"population_size" default:"256"`
	MaxGenerations int           `yaml:"max_generations" default:"128"`
	MaxAge         time.Duration `yaml:"max_age" default:"5m"`
}

type GeneticModelFactory struct {
	GeneticModelConfig
}

func (m *GeneticModelFactory) Config() interface{} {
	return &m.GeneticModelConfig
}

func (m *GeneticModelFactory) Create(orders, matchedOrders []*MarketOrder, log *zap.SugaredLogger) OptimizationMethod {
	return &GeneticModel{
		NewGenomeLab:   m.Genome.NewGenomeLab,
		PopulationSize: m.PopulationSize,
		MaxGenerations: m.MaxGenerations,
		MaxAge:         m.MaxAge,
		Log:            log.With(zap.String("model", "GMP")),
	}
}

type GeneticModel struct {
	NewGenomeLab NewGenomeLab
	// PopulationSize specifies the number of individuals per Population.
	PopulationSize int
	// MaxGenerations specifies the number of population generations.
	MaxGenerations int
	// MaxAge specifies the maximum age of the entire evolution process.
	MaxAge time.Duration
	// Log is used for internal logging.
	Log *zap.SugaredLogger
}

func (m *GeneticModel) ShouldEvolve(ga gago.GA) bool {
	return ga.Generations < m.MaxGenerations && ga.Age < m.MaxAge
}

func (m *GeneticModel) Optimize(knapsack *Knapsack, orders []*MarketOrder) error {
	if len(orders) == 0 {
		return fmt.Errorf("not enougn orders to perform optimization")
	}

	simulation := gago.Generational(m.NewGenomeLab(knapsack, orders))
	simulation.PopSize = m.PopulationSize
	simulation.Callback = func(ga *gago.GA) {
		if ga.Generations%(m.MaxGenerations/10) == 0 {
			m.Log.Debugf("optimization progress %3d/%3d, best price so far: %.12f", ga.Generations, m.MaxGenerations, -ga.HallOfFame[0].Fitness)
		}
	}
	simulation.Initialize()

	for m.ShouldEvolve(simulation) {
		simulation.Evolve()
	}

	survived := simulation.HallOfFame[0]

	if survived.Fitness == 0.0 {
		return fmt.Errorf("failed to evolute in %d generations", simulation.Generations)
	}

	winnersT, ok := survived.Genome.(interface {
		Pack(*Knapsack) error
	})
	if !ok {
		panic("genome must implement Pack method")
	}

	if err := winnersT.Pack(knapsack); err != nil {
		m.Log.Errorf("something wrong happened with genetic model: %v", err)
		return err
	}

	return nil
}
