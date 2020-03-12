package tables

type Predictor interface{
	Predict(*Table)*Table
	BatchSize() (int,int)
}

