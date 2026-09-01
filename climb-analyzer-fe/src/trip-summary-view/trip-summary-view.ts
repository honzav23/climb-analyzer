import {Component, input, OnChanges, output, SimpleChanges} from '@angular/core';
import { ElevationGraph } from '../elevation-graph/elevation-graph';
import { MapWrapper } from "../map-wrapper/map-wrapper";
import type {TripSummary, Climb} from '../types/AnalysisResponse';
import { DividerModule } from 'primeng/divider'

@Component({
  selector: 'trip-summary-view',
  imports: [ElevationGraph, DividerModule, MapWrapper],
  templateUrl: './trip-summary-view.html',
  styleUrl: './trip-summary-view.css'
})
export class TripSummaryView implements OnChanges {
    tripSummary = input<TripSummary | null>(null)
    climbs = input<Climb[]>([])
    summary: TripSummary | null = null
    selectedClimb = output<number>();

    ngOnChanges(changes: SimpleChanges) {
        if (changes['tripSummary'] && changes['tripSummary'].currentValue) {
            this.summary = changes['tripSummary'].currentValue;
    }
  }

}
