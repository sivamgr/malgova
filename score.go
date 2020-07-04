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

type AlgoScore struct {
	algoName string
	symbol   string
	// stats and scores
	ordersCount          int
	tradesCount          int
	tradesWon            int
	tradesLost           int
	winStreak            int
	lossStreak           int
	netPnl               float64
	netPnlPercentAverage float64
	netPnlPercentStdDev  float64

	sqn float64
}

func (t AlgoScore) String() string {
	return fmt.Sprintf("%12s|%20s|%5d|%4d|%4d:%4d|%3d:%3d| %9.2f |%9.2f|%9.2f| %7.3f", t.algoName, t.symbol, t.ordersCount, t.tradesCount, t.tradesWon, t.tradesLost, t.winStreak, t.lossStreak, t.netPnl, t.netPnlPercentAverage, t.netPnlPercentStdDev, t.sqn)
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
		algoName: a.algoName,
		symbol:   a.symbol,
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

	a.score.tradesCount = len(a.trades)
	pnl := make([]float64, 0)
	winStreak := 0
	lossStreak := 0

	if a.score.tradesCount > 0 {
		for _, t := range a.trades {
			a.score.ordersCount += t.orders
			if t.pnl > 0 {
				winStreak++
				lossStreak = 0
				a.score.tradesWon++
			} else {
				winStreak = 0
				lossStreak++
				a.score.tradesLost++
			}
			a.score.netPnl += t.pnl
			pnl = append(pnl, t.pnlPercentage)
			if a.score.winStreak < winStreak {
				a.score.winStreak = winStreak
			}
			if a.score.lossStreak < lossStreak {
				a.score.lossStreak = lossStreak
			}
		}
		a.score.netPnlPercentAverage = stat.Mean(pnl, nil)
		a.score.netPnlPercentStdDev = stat.StdDev(pnl, nil)
		if a.score.netPnlPercentStdDev != 0 {
			a.score.sqn = math.Sqrt(float64(a.score.tradesCount)) * a.score.netPnlPercentAverage / a.score.netPnlPercentStdDev
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
