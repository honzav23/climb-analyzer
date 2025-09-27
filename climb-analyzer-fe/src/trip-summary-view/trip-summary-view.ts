import {Component, input, OnChanges, SimpleChanges} from '@angular/core';
import { ElevationGraph } from '../elevation-graph/elevation-graph';
import type {TripSummary} from '../types/AnalysisResponse';

@Component({
  selector: 'trip-summary-view',
  imports: [ElevationGraph],
  templateUrl: './trip-summary-view.html',
  styleUrl: './trip-summary-view.css'
})
export class TripSummaryView implements OnChanges {
  tripSummary = input<TripSummary | null>(null)
  summary: TripSummary | null = null
  ngOnChanges(changes: SimpleChanges): void {
    // Check if the tripSummary input property was changed
    if (changes['tripSummary'] && changes['tripSummary'].currentValue) {
      // ALWAYS update your internal property here
      this.summary = changes['tripSummary'].currentValue;
    }
  }

}
