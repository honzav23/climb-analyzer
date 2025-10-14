import {Component, input, OnChanges, SimpleChanges} from '@angular/core';
import { ElevationGraph } from '../elevation-graph/elevation-graph';
import { MapWrapper } from "../map-wrapper/map-wrapper";
import type {TripSummary} from '../types/AnalysisResponse';
import { DividerModule } from 'primeng/divider'

@Component({
  selector: 'trip-summary-view',
  imports: [ElevationGraph, DividerModule, MapWrapper],
  templateUrl: './trip-summary-view.html',
  styleUrl: './trip-summary-view.css'
})
export class TripSummaryView implements OnChanges {
    tripSummary = input<TripSummary | null>(null)
    summary: TripSummary | null = null

    ngOnChanges(changes: SimpleChanges) {
        if (changes['tripSummary'] && changes['tripSummary'].currentValue) {
            this.summary = changes['tripSummary'].currentValue;
    }
  }

}
