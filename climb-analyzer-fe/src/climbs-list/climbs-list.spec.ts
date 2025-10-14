import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ClimbsList } from './climbs-list';

describe('ClimbsList', () => {
  let component: ClimbsList;
  let fixture: ComponentFixture<ClimbsList>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ClimbsList]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ClimbsList);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
