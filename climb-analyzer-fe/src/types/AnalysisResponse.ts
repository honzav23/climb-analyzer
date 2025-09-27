export type AnalysisResponse = {
  tripSummary: TripSummary
}

export type TripSummary = {
  lengthKilometers: number
  elevationGain: number
  elevationProfile: ElevationProfilePlotData[]
}

export type ElevationProfilePlotData = {
  distance: number
  elevation: number
}
