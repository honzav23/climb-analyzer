export type AnalysisResponse = {
  tripSummary: TripSummary;
  foundClimbs: Climb[]
}

export type PointCoordinates = {
    latitude: number
    longitude: number
}

export type TripSummary = {
  lengthKilometers: number;
  elevationGain: number;
  elevationProfile: ElevationProfilePlotData[];
  tripCoordinates: PointCoordinates[]
}

export type ElevationProfilePlotData = {
  distance: number;
  elevation: number;
}

export type Climb = {
    elevationGain: number;
    climbSegments: ClimbSegment[];
    length: number;
    averageGradient: number;
    start: number;
    end: number;
}

export type ClimbSegment = {
    elevationProfile: ElevationProfilePlotData[];
    averageGradient: number;
    segmentLength: number;
}
