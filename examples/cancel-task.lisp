(defmessage getTaskForKey
  (field :name key))

(defmessage taskForKey
  (field :name task))

(defmessage noTaskForKey)

(defprocess DynamoDBProcess
  (let ((tasksByKey {}))

    :start
    (loop
      (select
        (let (({key} (?receive :message getTaskForKey)))
          (if (map-contains? tasksByKey key)
              (!send :message getTaskForKey :task (map-get tasksByKey key))
              (!send :message noTaskForKey)))))))
