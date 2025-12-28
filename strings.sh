# Extract keys from source code, including T, Errorf, and Sprintf
grep -rho -E '\bt\.(Errorf|T|Sprintf)\("([^"\\]|\\.)*"' ./src --exclude='*_test.go' \
  | sed -E 's/t\.(Errorf|T|Sprintf)\("(([^"\\]|\\.)*)".*/\2/' \
  | sed -E 's/\\"/"/g' \
  | sort | uniq > keys.txt

if [ ! -f "keys.txt" ]; then
  echo "Error: keys.txt not found!"
  exit 1
fi

if [ ! -f "strings_de.ini" ]; then
  awk '{print $0"="}' "keys.txt" | sort > "strings_de.ini"
  echo "Created strings_de.ini with all keys."
else
  # Use awk to merge keys.txt and strings_de.ini
  # - If key is in keys.txt AND strings_de.ini: Keep as is (ensuring no semicolon)
  # - If key is in strings_de.ini but NOT in keys.txt: Comment out (prefix with ;)
  # - If key is in keys.txt but NOT in strings_de.ini: Add (with empty value)
  
  awk -F= '
    # Load keys from keys.txt (currently used)
    NR == FNR { used[$0] = 1; next }
    
    # Process strings_de.ini
    {
      line = $0
      # Extract key, handling potentially commented lines
      k = $1
      commented = (substr(k, 1, 1) == ";")
      if (commented) {
        real_key = substr(k, 2)
      } else {
        real_key = k
      }
      
      if (real_key in used) {
        # Key is used: uncomment if it was commented, keep as is otherwise
        print real_key "=" $2
        delete used[real_key]
      } else {
        # Key is NOT used: comment it out if not already
        if (commented) {
          print line
        } else {
          print ";" line
        }
      }
    }
    
    # Add completely new keys
    END {
      for (k in used) {
        print k "="
      }
    }
  ' keys.txt strings_de.ini | sort > "strings_de.ini.tmp" && mv "strings_de.ini.tmp" "strings_de.ini"

  echo "Updated strings_de.ini: added new keys and commented out unused ones."
fi

rm keys.txt