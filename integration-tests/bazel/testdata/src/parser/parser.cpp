#include "parser.h"

#include <cstdlib>
#include <limits>
#include <string>

#include "secrets.h"

int parse(const std::string &input) {
  if (input.empty()) {
    return -1;
  }
  if (input[0] == 'a' && input[1] == 'b' && input[2] == 'c') {
    if (input.find(SECRET_VALUE) != std::string::npos) {
      char *some_string = static_cast<char *>(malloc(4));
      free(some_string);
      // Crashes with AddressSanitizer, but should not crash without it: The
      // allocated memory is addressable, but has been freed before the access.
      return some_string[1];
    } else {
      // Trigger the UndefinedBehaviorSanitizer
      int n = 23;
      n <<= 32;
    }
  }
  return 0;
}
