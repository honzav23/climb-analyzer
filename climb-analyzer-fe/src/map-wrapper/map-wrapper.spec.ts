import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MapWrapper } from './map-wrapper';

describe('MapWrapper', () => {
  let component: MapWrapper;
  let fixture: ComponentFixture<MapWrapper>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MapWrapper]
    })
    .compileComponents();

    fixture = TestBed.createComponent(MapWrapper);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
