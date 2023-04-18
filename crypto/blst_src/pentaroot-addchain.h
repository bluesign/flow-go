/*
 * Copyright Supranational LLC
 * Licensed under the Apache License, Version 2.0, see LICENSE for details.
 * SPDX-License-Identifier: Apache-2.0
 */
/*
 * The "magic" number is 1/5 modulo BLS12_381_r-1. Exponentiation to which
 * yields 5th root of the base.
 *
 * Generated with 'addchain 20974350070050476191779096203274386335076221000211055129041463479975432473805'
 * https://github.com/kwantam/addchain
 * # Bos-Coster (win=4)           :  307 (15)
 * # Bos-Coster (win=10)          :  307 (18)
 * # Yacobi                       :  319 (16)
 * # Bos-Coster (win=2)           :  319 ( 5)
 * # Bos-Coster (win=5)           :  306 (19) <<<
 * # Bos-Coster (win=7)           :  311 (22)
 * # Bos-Coster (win=9)           :  313 (20)
 * # Bos-Coster (win=3)           :  314 ( 9)
 * # Bos-Coster (win=6)           :  309 (21)
 * # Bos-Coster (win=8)           :  309 (23)
 * # Bergeron-Berstel-Brlek-Duboc :  334 ( 5)
 */

#define PENTAROOT_MOD_BLS12_381_r(out, inp, ptype) do { \
ptype t[19]; \
vec_copy(t[1], inp, sizeof(ptype)); /*    0: 1 */\
sqr(t[7], t[1]);                    /*    1: 2 */\
sqr(t[0], t[7]);                    /*    2: 4 */\
sqr(t[2], t[0]);                    /*    3: 8 */\
mul(t[10], t[2], t[1]);             /*    4: 9 */\
mul(t[3], t[10], t[7]);             /*    5: b */\
mul(t[1], t[10], t[0]);             /*    6: d */\
mul(t[5], t[3], t[0]);              /*    7: f */\
mul(t[9], t[10], t[2]);             /*    8: 11 */\
mul(t[4], t[3], t[2]);              /*    9: 13 */\
mul(t[15], t[5], t[2]);             /*   10: 17 */\
mul(t[8], t[15], t[2]);             /*   11: 1f */\
mul(t[13], t[8], t[7]);             /*   12: 21 */\
mul(t[14], t[8], t[0]);             /*   13: 23 */\
mul(t[12], t[13], t[0]);            /*   14: 25 */\
mul(t[6], t[8], t[2]);              /*   15: 27 */\
mul(t[11], t[14], t[2]);            /*   16: 2b */\
sqr(t[0], t[15]);                   /*   17: 2e */\
mul(t[18], t[6], t[2]);             /*   18: 2f */\
mul(t[2], t[11], t[2]);             /*   19: 33 */\
mul(t[16], t[2], t[7]);             /*   20: 35 */\
mul(t[7], t[0], t[3]);              /*   21: 39 */\
mul(t[17], t[0], t[5]);             /*   22: 3d */\
/* sqr(t[0], t[0]); */              /*   23: 5c */\
/* sqr(t[0], t[0]); */              /*   24: b8 */\
/* sqr(t[0], t[0]); */              /*   25: 170 */\
/* sqr(t[0], t[0]); */              /*   26: 2e0 */\
/* sqr(t[0], t[0]); */              /*   27: 5c0 */\
/* sqr(t[0], t[0]); */              /*   28: b80 */\
/* sqr(t[0], t[0]); */              /*   29: 1700 */\
sqr_n_mul(t[0], t[0], 7, t[18]);    /*   30: 172f */\
/* sqr(t[0], t[0]); */              /*   31: 2e5e */\
/* sqr(t[0], t[0]); */              /*   32: 5cbc */\
/* sqr(t[0], t[0]); */              /*   33: b978 */\
/* sqr(t[0], t[0]); */              /*   34: 172f0 */\
/* sqr(t[0], t[0]); */              /*   35: 2e5e0 */\
/* sqr(t[0], t[0]); */              /*   36: 5cbc0 */\
sqr_n_mul(t[0], t[0], 6, t[13]);    /*   37: 5cbe1 */\
/* sqr(t[0], t[0]); */              /*   38: b97c2 */\
/* sqr(t[0], t[0]); */              /*   39: 172f84 */\
/* sqr(t[0], t[0]); */              /*   40: 2e5f08 */\
/* sqr(t[0], t[0]); */              /*   41: 5cbe10 */\
/* sqr(t[0], t[0]); */              /*   42: b97c20 */\
/* sqr(t[0], t[0]); */              /*   43: 172f840 */\
sqr_n_mul(t[0], t[0], 6, t[17]);    /*   44: 172f87d */\
/* sqr(t[0], t[0]); */              /*   45: 2e5f0fa */\
/* sqr(t[0], t[0]); */              /*   46: 5cbe1f4 */\
/* sqr(t[0], t[0]); */              /*   47: b97c3e8 */\
/* sqr(t[0], t[0]); */              /*   48: 172f87d0 */\
/* sqr(t[0], t[0]); */              /*   49: 2e5f0fa0 */\
/* sqr(t[0], t[0]); */              /*   50: 5cbe1f40 */\
sqr_n_mul(t[0], t[0], 6, t[16]);    /*   51: 5cbe1f75 */\
/* sqr(t[0], t[0]); */              /*   52: b97c3eea */\
/* sqr(t[0], t[0]); */              /*   53: 172f87dd4 */\
/* sqr(t[0], t[0]); */              /*   54: 2e5f0fba8 */\
/* sqr(t[0], t[0]); */              /*   55: 5cbe1f750 */\
/* sqr(t[0], t[0]); */              /*   56: b97c3eea0 */\
sqr_n_mul(t[0], t[0], 5, t[15]);    /*   57: b97c3eeb7 */\
/* sqr(t[0], t[0]); */              /*   58: 172f87dd6e */\
/* sqr(t[0], t[0]); */              /*   59: 2e5f0fbadc */\
/* sqr(t[0], t[0]); */              /*   60: 5cbe1f75b8 */\
/* sqr(t[0], t[0]); */              /*   61: b97c3eeb70 */\
/* sqr(t[0], t[0]); */              /*   62: 172f87dd6e0 */\
/* sqr(t[0], t[0]); */              /*   63: 2e5f0fbadc0 */\
sqr_n_mul(t[0], t[0], 6, t[15]);    /*   64: 2e5f0fbadd7 */\
/* sqr(t[0], t[0]); */              /*   65: 5cbe1f75bae */\
/* sqr(t[0], t[0]); */              /*   66: b97c3eeb75c */\
/* sqr(t[0], t[0]); */              /*   67: 172f87dd6eb8 */\
/* sqr(t[0], t[0]); */              /*   68: 2e5f0fbadd70 */\
/* sqr(t[0], t[0]); */              /*   69: 5cbe1f75bae0 */\
/* sqr(t[0], t[0]); */              /*   70: b97c3eeb75c0 */\
/* sqr(t[0], t[0]); */              /*   71: 172f87dd6eb80 */\
/* sqr(t[0], t[0]); */              /*   72: 2e5f0fbadd700 */\
sqr_n_mul(t[0], t[0], 8, t[14]);    /*   73: 2e5f0fbadd723 */\
/* sqr(t[0], t[0]); */              /*   74: 5cbe1f75bae46 */\
/* sqr(t[0], t[0]); */              /*   75: b97c3eeb75c8c */\
/* sqr(t[0], t[0]); */              /*   76: 172f87dd6eb918 */\
/* sqr(t[0], t[0]); */              /*   77: 2e5f0fbadd7230 */\
/* sqr(t[0], t[0]); */              /*   78: 5cbe1f75bae460 */\
/* sqr(t[0], t[0]); */              /*   79: b97c3eeb75c8c0 */\
/* sqr(t[0], t[0]); */              /*   80: 172f87dd6eb9180 */\
/* sqr(t[0], t[0]); */              /*   81: 2e5f0fbadd72300 */\
sqr_n_mul(t[0], t[0], 8, t[13]);    /*   82: 2e5f0fbadd72321 */\
/* sqr(t[0], t[0]); */              /*   83: 5cbe1f75bae4642 */\
/* sqr(t[0], t[0]); */              /*   84: b97c3eeb75c8c84 */\
/* sqr(t[0], t[0]); */              /*   85: 172f87dd6eb91908 */\
/* sqr(t[0], t[0]); */              /*   86: 2e5f0fbadd723210 */\
/* sqr(t[0], t[0]); */              /*   87: 5cbe1f75bae46420 */\
/* sqr(t[0], t[0]); */              /*   88: b97c3eeb75c8c840 */\
sqr_n_mul(t[0], t[0], 6, t[2]);     /*   89: b97c3eeb75c8c873 */\
/* sqr(t[0], t[0]); */              /*   90: 172f87dd6eb9190e6 */\
/* sqr(t[0], t[0]); */              /*   91: 2e5f0fbadd72321cc */\
/* sqr(t[0], t[0]); */              /*   92: 5cbe1f75bae464398 */\
/* sqr(t[0], t[0]); */              /*   93: b97c3eeb75c8c8730 */\
/* sqr(t[0], t[0]); */              /*   94: 172f87dd6eb9190e60 */\
/* sqr(t[0], t[0]); */              /*   95: 2e5f0fbadd72321cc0 */\
sqr_n_mul(t[0], t[0], 6, t[13]);    /*   96: 2e5f0fbadd72321ce1 */\
/* sqr(t[0], t[0]); */              /*   97: 5cbe1f75bae46439c2 */\
/* sqr(t[0], t[0]); */              /*   98: b97c3eeb75c8c87384 */\
/* sqr(t[0], t[0]); */              /*   99: 172f87dd6eb9190e708 */\
/* sqr(t[0], t[0]); */              /*  100: 2e5f0fbadd72321ce10 */\
/* sqr(t[0], t[0]); */              /*  101: 5cbe1f75bae46439c20 */\
/* sqr(t[0], t[0]); */              /*  102: b97c3eeb75c8c873840 */\
/* sqr(t[0], t[0]); */              /*  103: 172f87dd6eb9190e7080 */\
sqr_n_mul(t[0], t[0], 7, t[12]);    /*  104: 172f87dd6eb9190e70a5 */\
/* sqr(t[0], t[0]); */              /*  105: 2e5f0fbadd72321ce14a */\
/* sqr(t[0], t[0]); */              /*  106: 5cbe1f75bae46439c294 */\
/* sqr(t[0], t[0]); */              /*  107: b97c3eeb75c8c8738528 */\
/* sqr(t[0], t[0]); */              /*  108: 172f87dd6eb9190e70a50 */\
/* sqr(t[0], t[0]); */              /*  109: 2e5f0fbadd72321ce14a0 */\
/* sqr(t[0], t[0]); */              /*  110: 5cbe1f75bae46439c2940 */\
/* sqr(t[0], t[0]); */              /*  111: b97c3eeb75c8c87385280 */\
/* sqr(t[0], t[0]); */              /*  112: 172f87dd6eb9190e70a500 */\
sqr_n_mul(t[0], t[0], 8, t[11]);    /*  113: 172f87dd6eb9190e70a52b */\
/* sqr(t[0], t[0]); */              /*  114: 2e5f0fbadd72321ce14a56 */\
/* sqr(t[0], t[0]); */              /*  115: 5cbe1f75bae46439c294ac */\
/* sqr(t[0], t[0]); */              /*  116: b97c3eeb75c8c873852958 */\
/* sqr(t[0], t[0]); */              /*  117: 172f87dd6eb9190e70a52b0 */\
/* sqr(t[0], t[0]); */              /*  118: 2e5f0fbadd72321ce14a560 */\
/* sqr(t[0], t[0]); */              /*  119: 5cbe1f75bae46439c294ac0 */\
sqr_n_mul(t[0], t[0], 6, t[1]);     /*  120: 5cbe1f75bae46439c294acd */\
/* sqr(t[0], t[0]); */              /*  121: b97c3eeb75c8c873852959a */\
/* sqr(t[0], t[0]); */              /*  122: 172f87dd6eb9190e70a52b34 */\
/* sqr(t[0], t[0]); */              /*  123: 2e5f0fbadd72321ce14a5668 */\
/* sqr(t[0], t[0]); */              /*  124: 5cbe1f75bae46439c294acd0 */\
/* sqr(t[0], t[0]); */              /*  125: b97c3eeb75c8c873852959a0 */\
/* sqr(t[0], t[0]); */              /*  126: 172f87dd6eb9190e70a52b340 */\
/* sqr(t[0], t[0]); */              /*  127: 2e5f0fbadd72321ce14a56680 */\
/* sqr(t[0], t[0]); */              /*  128: 5cbe1f75bae46439c294acd00 */\
sqr_n_mul(t[0], t[0], 8, t[2]);     /*  129: 5cbe1f75bae46439c294acd33 */\
/* sqr(t[0], t[0]); */              /*  130: b97c3eeb75c8c873852959a66 */\
/* sqr(t[0], t[0]); */              /*  131: 172f87dd6eb9190e70a52b34cc */\
/* sqr(t[0], t[0]); */              /*  132: 2e5f0fbadd72321ce14a566998 */\
/* sqr(t[0], t[0]); */              /*  133: 5cbe1f75bae46439c294acd330 */\
/* sqr(t[0], t[0]); */              /*  134: b97c3eeb75c8c873852959a660 */\
/* sqr(t[0], t[0]); */              /*  135: 172f87dd6eb9190e70a52b34cc0 */\
sqr_n_mul(t[0], t[0], 6, t[11]);    /*  136: 172f87dd6eb9190e70a52b34ceb */\
/* sqr(t[0], t[0]); */              /*  137: 2e5f0fbadd72321ce14a56699d6 */\
/* sqr(t[0], t[0]); */              /*  138: 5cbe1f75bae46439c294acd33ac */\
/* sqr(t[0], t[0]); */              /*  139: b97c3eeb75c8c873852959a6758 */\
/* sqr(t[0], t[0]); */              /*  140: 172f87dd6eb9190e70a52b34ceb0 */\
sqr_n_mul(t[0], t[0], 4, t[10]);    /*  141: 172f87dd6eb9190e70a52b34ceb9 */\
/* sqr(t[0], t[0]); */              /*  142: 2e5f0fbadd72321ce14a56699d72 */\
/* sqr(t[0], t[0]); */              /*  143: 5cbe1f75bae46439c294acd33ae4 */\
/* sqr(t[0], t[0]); */              /*  144: b97c3eeb75c8c873852959a675c8 */\
/* sqr(t[0], t[0]); */              /*  145: 172f87dd6eb9190e70a52b34ceb90 */\
/* sqr(t[0], t[0]); */              /*  146: 2e5f0fbadd72321ce14a56699d720 */\
sqr_n_mul(t[0], t[0], 5, t[8]);     /*  147: 2e5f0fbadd72321ce14a56699d73f */\
/* sqr(t[0], t[0]); */              /*  148: 5cbe1f75bae46439c294acd33ae7e */\
/* sqr(t[0], t[0]); */              /*  149: b97c3eeb75c8c873852959a675cfc */\
/* sqr(t[0], t[0]); */              /*  150: 172f87dd6eb9190e70a52b34ceb9f8 */\
/* sqr(t[0], t[0]); */              /*  151: 2e5f0fbadd72321ce14a56699d73f0 */\
/* sqr(t[0], t[0]); */              /*  152: 5cbe1f75bae46439c294acd33ae7e0 */\
/* sqr(t[0], t[0]); */              /*  153: b97c3eeb75c8c873852959a675cfc0 */\
/* sqr(t[0], t[0]); */              /*  154: 172f87dd6eb9190e70a52b34ceb9f80 */\
/* sqr(t[0], t[0]); */              /*  155: 2e5f0fbadd72321ce14a56699d73f00 */\
/* sqr(t[0], t[0]); */              /*  156: 5cbe1f75bae46439c294acd33ae7e00 */\
/* sqr(t[0], t[0]); */              /*  157: b97c3eeb75c8c873852959a675cfc00 */\
/* sqr(t[0], t[0]); */              /*  158: 172f87dd6eb9190e70a52b34ceb9f800 */\
/* sqr(t[0], t[0]); */              /*  159: 2e5f0fbadd72321ce14a56699d73f000 */\
/* sqr(t[0], t[0]); */              /*  160: 5cbe1f75bae46439c294acd33ae7e000 */\
/* sqr(t[0], t[0]); */              /*  161: b97c3eeb75c8c873852959a675cfc000 */\
/* sqr(t[0], t[0]); */              /*  162: 172f87dd6eb9190e70a52b34ceb9f8000 */\
sqr_n_mul(t[0], t[0], 15, t[9]);    /*  163: 172f87dd6eb9190e70a52b34ceb9f8011 */\
/* sqr(t[0], t[0]); */              /*  164: 2e5f0fbadd72321ce14a56699d73f0022 */\
/* sqr(t[0], t[0]); */              /*  165: 5cbe1f75bae46439c294acd33ae7e0044 */\
/* sqr(t[0], t[0]); */              /*  166: b97c3eeb75c8c873852959a675cfc0088 */\
/* sqr(t[0], t[0]); */              /*  167: 172f87dd6eb9190e70a52b34ceb9f80110 */\
/* sqr(t[0], t[0]); */              /*  168: 2e5f0fbadd72321ce14a56699d73f00220 */\
/* sqr(t[0], t[0]); */              /*  169: 5cbe1f75bae46439c294acd33ae7e00440 */\
/* sqr(t[0], t[0]); */              /*  170: b97c3eeb75c8c873852959a675cfc00880 */\
/* sqr(t[0], t[0]); */              /*  171: 172f87dd6eb9190e70a52b34ceb9f801100 */\
sqr_n_mul(t[0], t[0], 8, t[3]);     /*  172: 172f87dd6eb9190e70a52b34ceb9f80110b */\
/* sqr(t[0], t[0]); */              /*  173: 2e5f0fbadd72321ce14a56699d73f002216 */\
/* sqr(t[0], t[0]); */              /*  174: 5cbe1f75bae46439c294acd33ae7e00442c */\
/* sqr(t[0], t[0]); */              /*  175: b97c3eeb75c8c873852959a675cfc008858 */\
/* sqr(t[0], t[0]); */              /*  176: 172f87dd6eb9190e70a52b34ceb9f80110b0 */\
/* sqr(t[0], t[0]); */              /*  177: 2e5f0fbadd72321ce14a56699d73f0022160 */\
sqr_n_mul(t[0], t[0], 5, t[8]);     /*  178: 2e5f0fbadd72321ce14a56699d73f002217f */\
/* sqr(t[0], t[0]); */              /*  179: 5cbe1f75bae46439c294acd33ae7e00442fe */\
/* sqr(t[0], t[0]); */              /*  180: b97c3eeb75c8c873852959a675cfc00885fc */\
/* sqr(t[0], t[0]); */              /*  181: 172f87dd6eb9190e70a52b34ceb9f80110bf8 */\
/* sqr(t[0], t[0]); */              /*  182: 2e5f0fbadd72321ce14a56699d73f002217f0 */\
/* sqr(t[0], t[0]); */              /*  183: 5cbe1f75bae46439c294acd33ae7e00442fe0 */\
/* sqr(t[0], t[0]); */              /*  184: b97c3eeb75c8c873852959a675cfc00885fc0 */\
/* sqr(t[0], t[0]); */              /*  185: 172f87dd6eb9190e70a52b34ceb9f80110bf80 */\
/* sqr(t[0], t[0]); */              /*  186: 2e5f0fbadd72321ce14a56699d73f002217f00 */\
/* sqr(t[0], t[0]); */              /*  187: 5cbe1f75bae46439c294acd33ae7e00442fe00 */\
/* sqr(t[0], t[0]); */              /*  188: b97c3eeb75c8c873852959a675cfc00885fc00 */\
sqr_n_mul(t[0], t[0], 10, t[7]);    /*  189: b97c3eeb75c8c873852959a675cfc00885fc39 */\
/* sqr(t[0], t[0]); */              /*  190: 172f87dd6eb9190e70a52b34ceb9f80110bf872 */\
/* sqr(t[0], t[0]); */              /*  191: 2e5f0fbadd72321ce14a56699d73f002217f0e4 */\
/* sqr(t[0], t[0]); */              /*  192: 5cbe1f75bae46439c294acd33ae7e00442fe1c8 */\
/* sqr(t[0], t[0]); */              /*  193: b97c3eeb75c8c873852959a675cfc00885fc390 */\
/* sqr(t[0], t[0]); */              /*  194: 172f87dd6eb9190e70a52b34ceb9f80110bf8720 */\
/* sqr(t[0], t[0]); */              /*  195: 2e5f0fbadd72321ce14a56699d73f002217f0e40 */\
sqr_n_mul(t[0], t[0], 6, t[6]);     /*  196: 2e5f0fbadd72321ce14a56699d73f002217f0e67 */\
/* sqr(t[0], t[0]); */              /*  197: 5cbe1f75bae46439c294acd33ae7e00442fe1cce */\
/* sqr(t[0], t[0]); */              /*  198: b97c3eeb75c8c873852959a675cfc00885fc399c */\
/* sqr(t[0], t[0]); */              /*  199: 172f87dd6eb9190e70a52b34ceb9f80110bf87338 */\
/* sqr(t[0], t[0]); */              /*  200: 2e5f0fbadd72321ce14a56699d73f002217f0e670 */\
/* sqr(t[0], t[0]); */              /*  201: 5cbe1f75bae46439c294acd33ae7e00442fe1cce0 */\
sqr_n_mul(t[0], t[0], 5, t[4]);     /*  202: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3 */\
/* sqr(t[0], t[0]); */              /*  203: b97c3eeb75c8c873852959a675cfc00885fc399e6 */\
/* sqr(t[0], t[0]); */              /*  204: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cc */\
/* sqr(t[0], t[0]); */              /*  205: 2e5f0fbadd72321ce14a56699d73f002217f0e6798 */\
/* sqr(t[0], t[0]); */              /*  206: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf30 */\
/* sqr(t[0], t[0]); */              /*  207: b97c3eeb75c8c873852959a675cfc00885fc399e60 */\
/* sqr(t[0], t[0]); */              /*  208: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cc0 */\
/* sqr(t[0], t[0]); */              /*  209: 2e5f0fbadd72321ce14a56699d73f002217f0e67980 */\
/* sqr(t[0], t[0]); */              /*  210: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf300 */\
sqr_n_mul(t[0], t[0], 8, t[2]);     /*  211: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf333 */\
/* sqr(t[0], t[0]); */              /*  212: b97c3eeb75c8c873852959a675cfc00885fc399e666 */\
/* sqr(t[0], t[0]); */              /*  213: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc */\
/* sqr(t[0], t[0]); */              /*  214: 2e5f0fbadd72321ce14a56699d73f002217f0e679998 */\
/* sqr(t[0], t[0]); */              /*  215: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3330 */\
/* sqr(t[0], t[0]); */              /*  216: b97c3eeb75c8c873852959a675cfc00885fc399e6660 */\
/* sqr(t[0], t[0]); */              /*  217: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc0 */\
/* sqr(t[0], t[0]); */              /*  218: 2e5f0fbadd72321ce14a56699d73f002217f0e6799980 */\
sqr_n_mul(t[0], t[0], 7, t[5]);     /*  219: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f */\
/* sqr(t[0], t[0]); */              /*  220: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e */\
/* sqr(t[0], t[0]); */              /*  221: b97c3eeb75c8c873852959a675cfc00885fc399e6663c */\
/* sqr(t[0], t[0]); */              /*  222: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78 */\
/* sqr(t[0], t[0]); */              /*  223: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f0 */\
/* sqr(t[0], t[0]); */              /*  224: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e0 */\
/* sqr(t[0], t[0]); */              /*  225: b97c3eeb75c8c873852959a675cfc00885fc399e6663c0 */\
/* sqr(t[0], t[0]); */              /*  226: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc780 */\
/* sqr(t[0], t[0]); */              /*  227: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f00 */\
/* sqr(t[0], t[0]); */              /*  228: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e00 */\
sqr_n_mul(t[0], t[0], 9, t[2]);     /*  229: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e33 */\
/* sqr(t[0], t[0]); */              /*  230: b97c3eeb75c8c873852959a675cfc00885fc399e6663c66 */\
/* sqr(t[0], t[0]); */              /*  231: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc */\
/* sqr(t[0], t[0]); */              /*  232: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f198 */\
/* sqr(t[0], t[0]); */              /*  233: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e330 */\
/* sqr(t[0], t[0]); */              /*  234: b97c3eeb75c8c873852959a675cfc00885fc399e6663c660 */\
/* sqr(t[0], t[0]); */              /*  235: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc0 */\
/* sqr(t[0], t[0]); */              /*  236: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f1980 */\
sqr_n_mul(t[0], t[0], 7, t[4]);     /*  237: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f1993 */\
/* sqr(t[0], t[0]); */              /*  238: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e3326 */\
/* sqr(t[0], t[0]); */              /*  239: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664c */\
/* sqr(t[0], t[0]); */              /*  240: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc98 */\
/* sqr(t[0], t[0]); */              /*  241: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19930 */\
/* sqr(t[0], t[0]); */              /*  242: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e33260 */\
/* sqr(t[0], t[0]); */              /*  243: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664c0 */\
/* sqr(t[0], t[0]); */              /*  244: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc980 */\
/* sqr(t[0], t[0]); */              /*  245: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f199300 */\
sqr_n_mul(t[0], t[0], 8, t[2]);     /*  246: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f199333 */\
/* sqr(t[0], t[0]); */              /*  247: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e332666 */\
/* sqr(t[0], t[0]); */              /*  248: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664ccc */\
/* sqr(t[0], t[0]); */              /*  249: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc9998 */\
/* sqr(t[0], t[0]); */              /*  250: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f1993330 */\
/* sqr(t[0], t[0]); */              /*  251: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e3326660 */\
/* sqr(t[0], t[0]); */              /*  252: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664ccc0 */\
/* sqr(t[0], t[0]); */              /*  253: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc99980 */\
/* sqr(t[0], t[0]); */              /*  254: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933300 */\
sqr_n_mul(t[0], t[0], 8, t[2]);     /*  255: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933333 */\
/* sqr(t[0], t[0]); */              /*  256: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e33266666 */\
/* sqr(t[0], t[0]); */              /*  257: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664ccccc */\
/* sqr(t[0], t[0]); */              /*  258: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc999998 */\
/* sqr(t[0], t[0]); */              /*  259: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f199333330 */\
/* sqr(t[0], t[0]); */              /*  260: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e332666660 */\
/* sqr(t[0], t[0]); */              /*  261: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664ccccc0 */\
/* sqr(t[0], t[0]); */              /*  262: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc9999980 */\
/* sqr(t[0], t[0]); */              /*  263: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f1993333300 */\
sqr_n_mul(t[0], t[0], 8, t[2]);     /*  264: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f1993333333 */\
/* sqr(t[0], t[0]); */              /*  265: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e3326666666 */\
/* sqr(t[0], t[0]); */              /*  266: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664ccccccc */\
/* sqr(t[0], t[0]); */              /*  267: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc99999998 */\
/* sqr(t[0], t[0]); */              /*  268: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933333330 */\
/* sqr(t[0], t[0]); */              /*  269: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e33266666660 */\
/* sqr(t[0], t[0]); */              /*  270: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664ccccccc0 */\
sqr_n_mul(t[0], t[0], 6, t[3]);     /*  271: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664cccccccb */\
/* sqr(t[0], t[0]); */              /*  272: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc999999996 */\
/* sqr(t[0], t[0]); */              /*  273: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933333332c */\
/* sqr(t[0], t[0]); */              /*  274: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e332666666658 */\
/* sqr(t[0], t[0]); */              /*  275: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664cccccccb0 */\
/* sqr(t[0], t[0]); */              /*  276: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc9999999960 */\
/* sqr(t[0], t[0]); */              /*  277: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933333332c0 */\
/* sqr(t[0], t[0]); */              /*  278: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e3326666666580 */\
/* sqr(t[0], t[0]); */              /*  279: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664cccccccb00 */\
sqr_n_mul(t[0], t[0], 8, t[2]);     /*  280: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664cccccccb33 */\
/* sqr(t[0], t[0]); */              /*  281: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc99999999666 */\
/* sqr(t[0], t[0]); */              /*  282: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933333332ccc */\
/* sqr(t[0], t[0]); */              /*  283: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e33266666665998 */\
/* sqr(t[0], t[0]); */              /*  284: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664cccccccb330 */\
/* sqr(t[0], t[0]); */              /*  285: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc999999996660 */\
/* sqr(t[0], t[0]); */              /*  286: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933333332ccc0 */\
/* sqr(t[0], t[0]); */              /*  287: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e332666666659980 */\
/* sqr(t[0], t[0]); */              /*  288: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664cccccccb3300 */\
sqr_n_mul(t[0], t[0], 8, t[2]);     /*  289: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664cccccccb3333 */\
/* sqr(t[0], t[0]); */              /*  290: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc9999999966666 */\
/* sqr(t[0], t[0]); */              /*  291: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933333332ccccc */\
/* sqr(t[0], t[0]); */              /*  292: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e3326666666599998 */\
/* sqr(t[0], t[0]); */              /*  293: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664cccccccb33330 */\
/* sqr(t[0], t[0]); */              /*  294: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc99999999666660 */\
/* sqr(t[0], t[0]); */              /*  295: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933333332ccccc0 */\
/* sqr(t[0], t[0]); */              /*  296: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e33266666665999980 */\
/* sqr(t[0], t[0]); */              /*  297: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664cccccccb333300 */\
sqr_n_mul(t[0], t[0], 8, t[2]);     /*  298: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664cccccccb333333 */\
/* sqr(t[0], t[0]); */              /*  299: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc999999996666666 */\
/* sqr(t[0], t[0]); */              /*  300: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933333332ccccccc */\
/* sqr(t[0], t[0]); */              /*  301: 5cbe1f75bae46439c294acd33ae7e00442fe1ccf3331e332666666659999998 */\
/* sqr(t[0], t[0]); */              /*  302: b97c3eeb75c8c873852959a675cfc00885fc399e6663c664cccccccb3333330 */\
/* sqr(t[0], t[0]); */              /*  303: 172f87dd6eb9190e70a52b34ceb9f80110bf8733cccc78cc9999999966666660 */\
/* sqr(t[0], t[0]); */              /*  304: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933333332ccccccc0 */\
sqr_n_mul(out, t[0], 6, t[1]);      /*  305: 2e5f0fbadd72321ce14a56699d73f002217f0e679998f19933333332cccccccd */\
} while(0)
