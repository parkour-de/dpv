LANGUAGES="de fr es tr ru ar pl ro ua sq"

# Extract keys from source code
# We focus on t.Errorf because t.T now takes an error (usually from t.Errorf)
grep -rho -E '\bt\.Errorf\("([^"\\]|\\.)*"' ./src --exclude='*_test.go' \
  | sed -E 's/t\.Errorf\("(([^"\\]|\\.)*)".*/\1/' \
  | sed -E 's/\\"/"/g' \
  | sort | uniq > keys.txt

if [ ! -f "keys.txt" ]; then
  echo "Error: keys.txt not found!"
  exit 1
fi

for LANG in $LANGUAGES; do
  FILE="strings_${LANG}.ini"
  
  if [ ! -f "$FILE" ]; then
    awk '{print $0"="}' "keys.txt" | sort > "$FILE"
    echo "Created $FILE with all keys."
  else
    # Use awk to merge keys.txt and existing ini
    awk -F= '
      # Load keys from keys.txt (currently used)
      NR == FNR { used[$0] = 1; next }
      
      # Process ini file
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
    ' keys.txt "$FILE" | sort > "${FILE}.tmp" && mv "${FILE}.tmp" "$FILE"
  
    echo "Updated $FILE: added new keys and commented out unused ones."
  fi
done

rm keys.txt