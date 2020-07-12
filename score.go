package malgova

import (
	"fmt"
	"math"
	"sort"

	"gonum.org/v1/gonum/stat"
)

type tradeData struct {
	algoName string
	symbol   string
	orders   []orderEntry
	score    AlgoScore
	trades   []tradeEntry
}

// AlgoScore struct
type AlgoScore struct {
	AlgoName string
	Symbol   string
	// stats and scores
	OrdersCount          int
	TradesCount          int
	TradesWon            int
	TradesLost           int
	WinStreak            int
	LossStreak           int
	NetPnl               float64
	NetPnlPercentAverage float64
	NetPnlPercentStdDev  float64

	SQN float64
}

func (t AlgoScore) String() string {
	return fmt.Sprintf("%12s|%20s|%5d|%4d|%4d:%4d|%3d:%3d| %9.2f |%9.2f|%9.2f| %7.3f", t.AlgoName, t.Symbol, t.OrdersCount, t.TradesCount, t.TradesWon, t.TradesLost, t.WinStreak, t.LossStreak, t.NetPnl, t.NetPnlPercentAverage, t.NetPnlPercentStdDev, t.SQN)
}

type tradeEntry struct {
	orders        int
	buyValue      float64
	sellValue     float64
	pnl           float64
	pnlPercentage float64
}

type algoTradeData struct {
	bySymbolTrades map[string]*tradeData
}

func (a *tradeData) add(t orderEntry) {
	a.orders = append(a.orders, t)
}

// reset score
func (a *tradeData) resetScore() {
	a.trades = make([]tradeEntry, 0)
	a.score = AlgoScore{
		AlgoName: a.algoName,
		Symbol:   a.symbol,
	}
}

func (a *tradeData) consolidateTrades() {
	//sort orders by time
	sort.Slice(a.orders, func(i, j int) bool {
		return a.orders[i].at.Before(a.orders[j].at)
	})

	// consolidate orders into trades
	pos := 0
	openTrade := tradeEntry{}
	for _, o := range a.orders {
		if pos == 0 {
			openTrade.orders = 0
			openTrade.buyValue = 0
			openTrade.sellValue = 0
		}
		pos += o.qty
		if o.qty > 0 {
			openTrade.buyValue = float64(o.qty) * o.price
		} else {
			openTrade.sellValue = -float64(o.qty) * o.price
		}
		openTrade.orders++

		if pos == 0 {
			openTrade.pnl = openTrade.sellValue - openTrade.buyValue
			if openTrade.buyValue > 0 {
				openTrade.pnlPercentage = (openTrade.pnl)
			} else if openTrade.pnl == 0 {
				openTrade.pnlPercentage = 0
			} else if openTrade.pnl < 0 {
				openTrade.pnlPercentage = -100
			} else {
				openTrade.pnlPercentage = 100
			}
			a.trades = append(a.trades, openTrade)
		}
	}
}

func (a *tradeData) processScore() {
	a.resetScore()
	a.consolidateTrades()

	a.score.TradesCount = len(a.trades)
	pnl := make([]float64, 0)
	winStreak := 0
	lossStreak := 0

	if a.score.TradesCount > 0 {
		for _, t := range a.trades {
			a.score.OrdersCount += t.orders
			if t.pnl > 0 {
				winStreak++
				lossStreak = 0
				a.score.TradesWon++
			} else {
				winStreak = 0
				lossStreak++
				a.score.TradesLost++
			}
			a.score.NetPnl += t.pnl
			pnl = append(pnl, t.pnlPercentage)
			if a.score.WinStreak < winStreak {
				a.score.WinStreak = winStreak
			}
			if a.score.LossStreak < lossStreak {
				a.score.LossStreak = lossStreak
			}
		}
		a.score.NetPnlPercentAverage = stat.Mean(pnl, nil)
		a.score.NetPnlPercentStdDev = stat.StdDev(pnl, nil)
		if a.score.NetPnlPercentStdDev != 0 {
			a.score.SQN = math.Sqrt(float64(a.score.TradesCount)) * a.score.NetPnlPercentAverage / a.score.NetPnlPercentStdDev
		}

	}
}

func calculateAlgoScores(orders []orderEntry) []AlgoScore {
	scores := make([]AlgoScore, 0)
	mapAlgoData := make(map[string]*algoTradeData)
	for _, t := range orders {
		if _, ok := mapAlgoData[t.algoName]; !ok {
			mapAlgoData[t.algoName] = new(algoTradeData)
			mapAlgoData[t.algoName].bySymbolTrades = make(map[string]*tradeData)
		}
		if _, ok := mapAlgoData[t.algoName].bySymbolTrades[t.symbol]; !ok {
			mapAlgoData[t.algoName].bySymbolTrades[t.symbol] = new(tradeData)
			mapAlgoData[t.algoName].bySymbolTrades[t.symbol].orders = make([]orderEntry, 0)
			mapAlgoData[t.algoName].bySymbolTrades[t.symbol].algoName = t.algoName
			mapAlgoData[t.algoName].bySymbolTrades[t.symbol].symbol = t.symbol
		}
		mapAlgoData[t.algoName].bySymbolTrades[t.symbol].add(t)
	}

	for _, a := range mapAlgoData {
		for _, st := range a.bySymbolTrades {
			st.processScore()
			scores = append(scores, st.score)
		}
	}

	return scores
}
