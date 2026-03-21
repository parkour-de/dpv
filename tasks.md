# DPV UAT Feedback & Tasks

## 1. Kündigungsdatum soll nicht frei wählbar sein
**Rule:** Satzung: "Der Austritt kann zum Ende eines Kalenderjahres unter Einhaltung einer Frist von 3 Monaten erklärt werden"
*   **Backend:** Remove `CancellationDate` from incoming payload and `api.raml`. The backend calculates the cancellation date automatically based on the rule (min 3 months notice to end of year).
*   **Frontend:** Remove field. Create a copy of the logic to show a message: "if you cancel now, your membership will end on [Date]".

## 2. Eintrittsdatum legt der Admin fest
*   **Membership fields:** Add `ApplicationDate`, `BeginDate`, `EndDate` to the model. Update `api.raml` and `types.ts`.
*   **Backend:** `ApplicationDate` is automatically filled by backend. Remove it from API parameters.
*   **Frontend Admin:** `BeginDate` can be changed by Admin in a modal when accepting a membership.
*   **Frontend Overview:** For unconfirmed memberships, show `ApplicationDate` instead of `BeginDate` using an adaptive label ("Antragsdatum" / "Eintrittsdatum").

## 3. Löschen von Vereinen nicht möglich, wenn Mitgliedschaft vorliegt
*   **Backend:** Prevent deletion of a club if an active membership exists.
*   **Frontend:** Hide the delete button if active membership exists.

## 4. Bug: Kontoinhaber wird nicht angezeigt
*   **Frontend/Backend:** Ensure `AccountHolder` and `SEPAMandateNumber` are rendered correctly and populated with data for the user themselves, a club Vorstand user, and any admin.

## 5. Wunsch: Name, Rechtsform und Registernummer nach der Antragstellung nur noch von Admins bearbeitbar
*   **Frontend/Backend:** If membership status is anything but the initial value (e.g. inactive or empty string), club owners can no longer change `Name`, `LegalForm` (Rechtsform), and `RegistrationNumber` (Registernummer). Only admins can.

## 6. Bug: Werte auf 0 setzen und leere Felder werden nicht an den Server weitergeleitet
*   **Backend Partial Updates:** Go treats empty values as default. Since `map[string]interface{}` is used for partial updates, ensure clearing optional fields (like address) works.
*   **Validation:** Missing minimum mandatory fields (`Name` for Club; `FirstName`, `LastName`, `Email` for User) should return a validation error instead of clearing them: "field %s must not be empty".

## 7. Bestandsmeldung --> im Vorstand klären, bis wann wir die Deadlines setzen wollen
*   **Status:** *IGNORE - DO NOT WORK ON THIS.* Tracked here but not specified enough.

## 8. Bug: Vorstand nach 26BGB aktuell nur von Admins setzbar
*   **Frontend/Backend:** Anyone with basic edit rights on clubs (existing Vorstand and Admin) should see the edit button next to the Vorstand list to enable/disable the `AuthorizedRepresentative` flag.

## 9. Feedback für das Präsidium: Wer darf wann was --> Prozessabläufe spezifizieren
*   **Status:** *IGNORE - DO NOT WORK ON THIS.* Tracked here but not specified enough.

## 10. Admin Routine: Show clubs lacking Census
*   **Feature:** Add a feature for Admins to show all clubs that have a membership state and haven't uploaded a Census for the current year.

## 11. Repeating Task & Membership Status Logic
*   **Repeating Task:** Trigger ~1 minute after server start, and at midnight.
*   **Logic:** Read `BeginDate` and `EndDate` of user and club memberships. Update statuses if they don't fit the dates.
*   **New Statuses:** 
    *   "approved but not active yet": behaves like not a member yet.
    *   "ending due to being in cancellation": behaves exactly as active.
*   **Status Transitions:** If someone cancels, status becomes "ending due to being in cancellation" rather than immediately cancelled.
*   **Constraints:** `BeginDate` and `EndDate` must NOT be set when someone applies. `EndDate` must NOT be set when an application is approved. Ensure existing modifying service functions adhere to this.
