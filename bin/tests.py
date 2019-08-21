#!/usr/bin/env python

import unittest as ut

import gamma as gm


class TestGamma(ut.TestCase):
    def test_point_int_rect(self):
        rect = gm.Rect(max_x=1.0, min_x=0.0, max_y=2.0, min_y=-1.0)
        point1 = gm.Point(0.5, -0.8)
        point2 = gm.Point(2.0, -0.8)
        
        self.assertTrue(gm.point_in_rect(point1, rect), "point1 should be in "
                        "rect")
        self.assertFalse(gm.point_in_rect(point2, rect), "point2 should not "
                         "be in rect")
    


def main():
    ut.main()


if __name__ == "__main__":
    main()